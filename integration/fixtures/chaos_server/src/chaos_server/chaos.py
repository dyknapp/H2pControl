from chaos_server_proto.chaos_server import ChaosServiceBase, DelayRequest, Empty

import logging
import time
import os
import asyncio
import faulthandler

class ChaosService(ChaosServiceBase):
    def __init__(self, args) -> None:
        self.logger = logging.getLogger(__name__)
        self.mode = args.mode or "normal"
        self.delay = 3.0 if args.delay is None else args.delay


    def activate_startup_mode(self) -> None:
        loop = asyncio.get_running_loop()

        if self.mode == "normal":
            self.logger.info("Prepared to cause chaos")
        elif self.mode == "crash_after":
            loop.call_later(self.delay, self.force_segfault)
        elif self.mode == "exit_after":
            loop.call_later(self.delay, lambda: os._exit(1))
        elif self.mode == "abort_after":
            loop.call_later(self.delay, os.abort)
        elif self.mode == "hang_for":
            loop.call_soon(self._hang_server, self.delay)
        else:
            raise ValueError(f"Invalid mode: {self.mode}")
    

    def force_segfault(self) -> None:
        sigsegv = getattr(faulthandler, "_sigsegv", None)
        if sigsegv is None:
            raise RuntimeError("This Python implementation cannot generate a segfault!")
        sigsegv()

    def _crash_server(self, delay: float) -> None:
        self.logger.info(f"intentionally crashing after delay of {delay}s")
        time.sleep(delay)
        self.force_segfault()
        self.logger.error("Hmmm... I should have crashed.")

    async def crash_server(self, message: DelayRequest) -> Empty:
        self._crash_server(message.delay)
        return Empty()


    

    def _exit_server(self, delay: float) -> None:
        self.logger.info(f"intentionally exiting after delay of {delay}s")
        time.sleep(delay)
        os._exit(1)
        self.logger.error("Hmmm... I should have exited.")

    async def exit_server(self, message: DelayRequest) -> Empty:
        self._exit_server(message.delay)
        return Empty()


    

    def _abort_server(self, delay: float) -> None:
        self.logger.info(f"intentionally aborting after delay of {delay}s")
        time.sleep(delay)
        os.abort()
        self.logger.error("Hmmm... I should have aborted.")

    async def abort_server(self, message: DelayRequest) -> Empty:
        self._abort_server(message.delay)
        return Empty()


    

    def _hang_server(self, delay: float) -> None:
        self.logger.info(f"intentionally hanging for {delay}s")
        time.sleep(delay)
        self.logger.info("done hanging")

    async def hang_server(self, message: DelayRequest) -> Empty:
        # NOT async!
        self._hang_server(message.delay)
        return Empty()
