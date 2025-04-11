import asyncio
from stubs.arduino.python.v1.arduino import GreeterBase, HelloReply, HelloRequest
from grpclib.server import Server

class GreetingService(GreeterBase):
  async def say_hello(self, message: HelloRequest) -> HelloReply:
    return HelloReply(message="World")

  async def say_hello_again(self, message: HelloRequest) -> HelloReply:
    return HelloReply(message="World again!")

async def main():
  server = Server([GreetingService()])
  await server.start("127.0.0.1", 50051)
  await server.wait_closed()
  
if __name__ == '__main__':
    asyncio.run(main())