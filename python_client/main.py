import asyncio
from stubs.arduino.python.v1.arduino import GreeterStub, HelloRequest
# from grpclib.client import Channel
import grpc
async def main():
    # Use async context manager for proper cleanup
  channel = grpc.insecure_channel("localhost:50051")
  service = GreeterStub(channel)
  response = service.say_hello(HelloRequest(name="Tijmen"))
  print(response.message)
      
if __name__ == "__main__":
      asyncio.run(main())