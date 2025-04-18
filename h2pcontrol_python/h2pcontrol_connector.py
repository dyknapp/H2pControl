import grpc
from pb.h2pcontrol import Empty, ManagerStub, FetchServersResponse
from typing import Type, Tuple, TypeVar

TStub = TypeVar("TStub")

class H2PControl:
    def __init__(self, address: str):
        self.address = address
        self.channel = None
        self.service = None

    async def connect(self):
        self.channel = grpc.aio.insecure_channel(self.address)
        self.service = ManagerStub(self.channel)
       

    async def register_server(
            self, name: str, stub: Type[TStub]
        ) -> Tuple[grpc.aio.Channel, TStub]:
            response: FetchServersResponse = await self.service.fetch_servers(Empty())
            for server in response.servers:
                if server.name == name:
                    channel = grpc.aio.insecure_channel(server.addr)
                    service = stub(channel)
                    return channel, service
            raise ValueError(f"Server named {name} not found")
      
    async def close(self):
        if self.channel:
            await self.channel.close()
            self.channel = None
            self.service = None

