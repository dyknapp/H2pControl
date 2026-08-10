import argparse
import asyncio
import logging
import signal
import sys

import tomllib
from grpclib.server import Server
from grpclib.health.service import Health


from chaos_server.chaos import ChaosService

async def main(args, port_override=None):
    health_service = Health()
    chaos_service = ChaosService(args)

    # Add to this list with your actual service implementations
    server = Server([health_service, chaos_service])

    # We gather the port from the h2pcontrol.server.toml file by default, if we can not get that port we take a default port.
    bind_host = configuration.get("bind_host", "127.0.0.1")
    port = int(port_override or configuration.get("port", 50052))
    await server.start(bind_host, port)

    logger.info(f"Server started on {bind_host}:{port}")

    chaos_service.activate_startup_mode()

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
    else:
        signal.signal(signal.SIGINT, lambda s, f: _signal_handler())
        if hasattr(signal, "SIGBREAK"):
            signal.signal(signal.SIGBREAK, lambda s, f: _signal_handler())

    await should_stop.wait()

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

    # Additional command line arguments
    parser.add_argument(
        "--mode",
        help = "normal     : wait for RPC (this is the default mode)\n" +
               "exit_after : exit  after delay (with --delay, 3s default)\n" +
               "abort_after: abort after delay (with --delay, 3s default)\n" +
               "crash_after: crash after delay (with --delay, 3s default)\n" +
               "hang_for   : hang until delay elapses (with --delay, 3s default)\n",
        type=str,
    )
    parser.add_argument(
        "--delay",
        help = "Set time delay/span",
        type=float
    )

    args = parser.parse_args()

    asyncio.run(main(args, args.port))
