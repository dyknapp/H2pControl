import asyncio
import logging
import signal
from stubs.arduino.python.v1.arduino import GreeterBase, HelloReply, HelloRequest
from grpclib.server import Server
import tomllib


# Could consider this boilerplate to be done by our h2p library
# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s [%(levelname)s] %(message)s',
    handlers=[logging.StreamHandler()]
)
logger = logging.getLogger(__name__)


# --- Read configuration from TOML ---
# We could consider putting this tomllib as a subdependency of our h2p library to make this all simpler
with open("h2pcontrol.server.toml", "rb") as f:
    config = tomllib.load(f)
configuration = config.get("configuration", {})
# Now you can use configuration["port"]

class GreetingService(GreeterBase):
    async def say_hello(self, message: HelloRequest) -> HelloReply:
        logger.info(f"Received request: {message}")
        response = HelloReply(message="World")
        logger.info(f"Sending response: {response}")
        return response

    async def say_hello_again(self, message: HelloRequest) -> HelloReply:
        logger.info(f"Received request: {message}")
        response = HelloReply(message="World again!")
        logger.info(f"Sending response: {response}")
        return response


# There is a lot of syntactic sugar here, could resolve this?
async def main():
    server = Server([GreetingService()])
    port = configuration.get("port", 50052)
    await server.start("127.0.0.1", port)
    logger.info(f"Server started on 127.0.0.1:{port}")

    # Use an asyncio Event to wait for shutdown signal
    should_stop = asyncio.Event()

    def _signal_handler():
        logger.info("Shutdown signal received.")
        should_stop.set()

    loop = asyncio.get_running_loop()
    for sig in (signal.SIGINT, signal.SIGTERM):
        loop.add_signal_handler(sig, _signal_handler)

    await should_stop.wait()
    logger.info("Shutting down server...")
    await server.close()
    await server.wait_closed()
    logger.info("Server shutdown complete.")

if __name__ == '__main__':
    asyncio.run(main())
