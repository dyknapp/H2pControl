import argparse
import asyncio
import logging
import signal
import sys

import tomllib
from grpclib.server import Server


async def main(port_override=None):
    # Replace this list with your actual service implementations
    server = Server([])

    # We gather the port from the h2pcontrol.server.toml file by default, if we can not get that port we take a default port.
    port = port_override or configuration.get("port", 50052)
    await server.start("127.0.0.1", port)

    logger.info(f"Server started on 127.0.0.1:{port}")

    # Use an asyncio Event to wait for shutdown signal
    should_stop = asyncio.Event()

    # To gracefully handle shutdown
    def _signal_handler():
        logger.info("Shutdown signal received.")
        should_stop.set()

    loop = asyncio.get_running_loop()
    if sys.platform != "win32":
        for sig in (signal.SIGINT, signal.SIGTERM):
            loop.add_signal_handler(sig, _signal_handler)
        await should_stop.wait()
    else:
        # Windows: Use synchronous signal.signal for SIGBREAK and SIGINT
        signal.signal(signal.SIGINT, _signal_handler)  # Ctrl+C
        if hasattr(signal, "SIGBREAK"):
            signal.signal(
                signal.SIGBREAK, _signal_handler
            )  # Ctrl+Break/CTRL_BREAK_EVENT

    logger.info("Shutting down server...")
    server.close()
    await server.wait_closed()
    logger.info("Server shutdown complete.")


# Default logging configuration
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    handlers=[logging.StreamHandler()],
)
logger = logging.getLogger(__name__)

with open("h2pcontrol.server.toml", "rb") as f:
    config = tomllib.load(f)
configuration = config.get("configuration", {})

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Start the gRPC server")
    parser.add_argument("--port", type=int, help="Port number to listen on")
    args = parser.parse_args()

    asyncio.run(main(args.port))
