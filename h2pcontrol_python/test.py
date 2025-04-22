import asyncio

import grpc
from h2pcontrol_connector import H2PControl
from hpcontrol_python_server.arduino import GreeterStub, HelloRequest


async def main():
    h2pcontroller = H2PControl("localhost:50051")
    await h2pcontroller.connect()
    
    print(h2pcontroller.servers)
    
    
    print(h2pcontroller.servers.arduino)

    servers = h2pcontroller.servers
    # Should have a h2pcontroller.servers and then u can do h2pcontroller.servers.arduino?
    
    channel, server = await h2pcontroller.register_server(servers.arduino, GreeterStub)
    
    print(await server.say_hello(HelloRequest()))
    # print(await service.say_hello(message=HelloRequest()))
    
    # print("kek")

    # Prepare several say_hello requests
    requests = [server.say_hello(HelloRequest(name=f"User {i}")) for i in range(1000)]
    
    # Await all at once
    responses = await asyncio.gather(*requests)
    
    # print("lmmao")
    await h2pcontroller.close()

if __name__ == "__main__":
    asyncio.run(main())


