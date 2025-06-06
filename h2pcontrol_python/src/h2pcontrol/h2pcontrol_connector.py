import csv
import functools
from enum import Enum
from typing import Any as TypingAny
from typing import Callable, Tuple, Type, TypeVar

import grpc
import influxdb_client
from influxdb_client import Point

from h2pcontrol.pb.h2pcontrol import (
    Empty,
    FetchServersResponse,
    ManagerStub,
)

TStub = TypeVar("TStub")


class ServerAccessor:
    def __init__(self, server_names):
        super().__init__()
        for name in server_names:
            # Only allow valid Python identifiers as attributes
            if name.isidentifier():
                setattr(self, name, name)


Mode = Enum("Mode", ["In", "Out"])

F = TypeVar("F", bound=Callable[..., TypingAny])


class DATASINK(Enum):
    CSV = "csv"
    INFLUX = "influx"


class H2PControl:
    def __init__(self, address: str):
        self.datasink = DATASINK.INFLUX
        self.address = address
        self.channel = None
        self.manager = None
        self.servers = []

    class H2PDataSink:
        def __init__(self, outer, *modes):
            self.outer = outer
            self.log_in = Mode.In in modes
            self.log_out = Mode.Out in modes

        def __call__(self, func):
            func_name = getattr(func, "__name__")

            @functools.wraps(func)
            def wrapper(*args, **kwargs):
                # Handle input arguments
                if self.log_in:
                    # Hardcoded for classes, #0 is self, #1 should be grpc value
                    msg_in = args[1]
                    self._handle_data(msg_in, func_name, direction="in")

                result = func(*args, **kwargs)

                # Handle return values
                if self.log_out:
                    self._handle_data(result, func_name, direction="out")

                return result

            return wrapper

        def _handle_data(self, msg, func_name, direction):
            if hasattr(msg, "to_dict"):
                flat = self.flatten_with_hierarchical_tags(msg.to_dict())
                self.output(flat, func_name, direction)
            else:
                print(f"[{direction}] {func_name}: Not a message or missing to_dict()")

        def flatten_with_hierarchical_tags(self, d, parent_keys=None):
            if parent_keys is None:
                parent_keys = []
            data_points = []
            for k, v in d.items():
                path = parent_keys + [k]
                if isinstance(v, dict) and v:
                    data_points.extend(self.flatten_with_hierarchical_tags(v, path))
                else:
                    tags = {key: "1" for key in parent_keys}
                    field = k
                    data_points.append((tags, field, v))
            return data_points

        def output(self, flat, func_name, direction):
            sink = getattr(self.outer, "datasink", DATASINK.CSV)
            if sink == DATASINK.CSV:
                self.to_csv(flat, func_name, direction)
            elif sink == DATASINK.INFLUX:
                self.to_influx(flat, func_name, direction)
            else:
                print(f"[{direction}] Unknown datasink, printing:")
                for tags, field, value in flat:
                    print(f"Tags: {tags}, Field: {field}, Value: {value}")

        def to_csv(self, flat, func_name, direction):
            row = {}
            for tags, field, value in flat:
                for tag_k, tag_v in tags.items():
                    row[f"tag_{tag_k}"] = tag_v
                row[field] = value
            filename = f"{func_name}_{direction}.csv"
            try:
                with open(filename, "a", newline="") as f:
                    writer = csv.DictWriter(f, fieldnames=row.keys())
                    if f.tell() == 0:
                        writer.writeheader()
                    writer.writerow(row)
                print(f"[{direction}] Wrote to {filename}: {row}")
            except Exception as e:
                print(f"[{direction}] CSV write error: {e}")

        def to_influx(self, flat, func_name, direction):
            influxdb_url = "http://localhost:8086"
            token = "#"
            org = "beyerlab"
            bucket = "test"
            client = influxdb_client.InfluxDBClient(
                url=influxdb_url, token=token, org=org
            )

            # Fix: Use SYNCHRONOUS write options instead of None
            write_api = client.write_api(
                write_options=influxdb_client.client.write_api.SYNCHRONOUS
            )

            for tags, field, value in flat:
                if value is None or isinstance(value, (dict, list)):
                    continue
                point = Point(f"{func_name}_{direction}")
                for tag_k, tag_v in tags.items():
                    point = point.tag(
                        tag_k, tag_v
                    )  # Fix: assign back the returned Point
                point = point.field(field, value)  # Fix: assign back the returned Point
                write_api.write(bucket=bucket, org=org, record=point)

            # Fix: Close client to prevent resource leaks
            client.close()
            print(f"[{direction}] Wrote to InfluxDB: {func_name}_{direction}")

    def data_sink(self, *modes: Mode):
        return self.H2PDataSink(self, *modes)

    async def connect(self):
        self.channel = grpc.aio.insecure_channel(self.address)
        self.manager = ManagerStub(self.channel)

        fetchServerResponse: FetchServersResponse = await self.manager.fetch_servers(
            Empty()
        )

        server_names = [server.name for server in fetchServerResponse.servers]
        self.servers = ServerAccessor(server_names)

    async def register_server(
        self, name: str, stub: Type[TStub]
    ) -> Tuple[grpc.aio.Channel, TStub]:
        response: FetchServersResponse = await self.manager.fetch_servers(Empty())
        for server in response.servers:
            if server.name == name:
                channel = grpc.aio.insecure_channel(server.addr)
                server = stub(channel)
                return channel, server
        raise ValueError(f"Server named {name} not found")

    async def close(self):
        if self.channel:
            await self.channel.close()
            self.channel = None
            self.manager = None
