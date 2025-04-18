import asyncio

import grpc
from h2pcontrol_connector import H2PControl
from hpcontrol_python_server.arduino import GreeterStub, HelloRequest


async def main():
    h2pcontroller = H2PControl("localhost:50051")
    await h2pcontroller.connect()
    
    # Should have a h2pcontroller.servers and then u can do h2pcontroller.servers.arduino?
    
    channel, service = await h2pcontroller.register_server("arduino", GreeterStub)
    print(await service.say_hello(message=HelloRequest()))
    
    response = await service.say_hello_again(HelloRequest())
    print(response.message)

    await h2pcontroller.close()

if __name__ == "__main__":
    asyncio.run(main())
