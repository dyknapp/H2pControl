import grpc
from pb.h2pcontrol import Empty, ManagerStub, FetchServersResponse, FetchServerDefinition
from typing import Dict, Type, Tuple, TypeVar
from functools import reduce

TStub = TypeVar("TStub")

class ServerAccessor:
    def __init__(self, server_names):
        super().__init__()
        for name in server_names:
            # Only allow valid Python identifiers as attributes
            if name.isidentifier():
                setattr(self, name, name)
    
class H2PControl:
    def __init__(self, address: str):
        self.address = address
        self.channel = None
        self.manager = None
        self.servers = []

    async def connect(self):
        self.channel = grpc.aio.insecure_channel(self.address)
        self.manager = ManagerStub(self.channel)
        
        fetchServerResponse : FetchServersResponse = await self.manager.fetch_servers(Empty())
        
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

