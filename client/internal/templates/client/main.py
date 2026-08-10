import asyncio

# Import the h2pcontrol client library.
from h2pcontrol.h2pcontrol_connector import H2PControl


async def main():
    # Connect to the H2P manager.
    h2pcontroller = H2PControl("localhost:50051")
    await h2pcontroller.connect()

    try:
        # Find available services with `h2pcontrol fetch`. After installing a
        # generated stub package, import its Stub class and register it here:
        #
        # _, service = await h2pcontroller.register_server("service-name", ServiceStub)
        # response = await service.some_method(SomeRequest())
        # print(response)
        pass
    finally:
        await h2pcontroller.close()


if __name__ == "__main__":
    asyncio.run(main())
