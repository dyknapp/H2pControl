import asyncio
import logging
import signal
from grpclib.server import Server
import tomllib

# Replace this with your service, e.g. HelloWorldService(HelloWorldBase)
class Service(Base):    
    # Implement your business logic here
    def __init__():
      raise NotImplementedError() 

async def main():
    # Replace this list with your actual service implementations
    server = Server([Service()])
    
    # We gather the port from the h2pcontrol.server.toml file by default, if we can not get that port we take a default port.
    port = configuration.get("port", 50052)
    await server.start("127.0.0.1", port)
    
    logger.info(f"Server started on 127.0.0.1:{port}")

    # Use an asyncio Event to wait for shutdown signal
    should_stop = asyncio.Event()
    
    # To gracefully handle shutdown
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

# Default logging configuration
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s [%(levelname)s] %(message)s',
    handlers=[logging.StreamHandler()]
)
logger = logging.getLogger(__name__)

# --- Read configuration from TOML ---
with open("h2pcontrol.server.toml", "rb") as f:
    config = tomllib.load(f)
configuration = config.get("configuration", {})

if __name__ == '__main__':
    asyncio.run(main())
