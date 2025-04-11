import asyncio
import logging
from stubs.arduino.python.v1.arduino import GreeterBase, HelloReply, HelloRequest
from grpclib.server import Server

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s [%(levelname)s] %(message)s',
    handlers=[logging.StreamHandler()]
)
logger = logging.getLogger(__name__)

class GreetingService(GreeterBase):
    async def say_hello(self, message: HelloRequest) -> HelloReply:
        # Print incoming request
        logger.info(f"Received request: {message}")
        
        response = HelloReply(message="World")
        
        # Print outgoing response
        logger.info(f"Sending response: {response}")
        return response

    async def say_hello_again(self, message: HelloRequest) -> HelloReply:
        # Print incoming request
        logger.info(f"Received request: {message}")
        
        response = HelloReply(message="World again!")
        
        # Print outgoing response
        logger.info(f"Sending response: {response}")
        return response

async def main():
    server = Server([GreetingService()])
    await server.start("127.0.0.1", 50052)
    logger.info("Server started on 127.0.0.1:50052")
    await server.wait_closed()
    logger.info("Server shutdown complete")

if __name__ == '__main__':
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        logger.info("Shutting down...")
