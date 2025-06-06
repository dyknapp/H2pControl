import asyncio

from h2pcontrol_connector import H2PControl, H2PDataSink, Mode


@H2PDataSink(Mode.In)
def add(a, b):
    return a + b


@H2PDataSink(Mode.In, Mode.Out)
def multiply(a, b):
    return a * b


@H2PDataSink(Mode.In, Mode.Out)
def subtract(a, b):
    return a - b


async def main():
    h2pcontroller = H2PControl("localhost:50051")
    await h2pcontroller.connect()

    print(h2pcontroller.servers)

    # print(h2pcontroller.servers.arduino)

    servers = h2pcontroller.servers
    # Should have a h2pcontroller.servers and then u can do h2pcontroller.servers.arduino?

    # channel, server = await h2pcontroller.register_server(servers.arduino, GreeterStub)

    # print(await server.say_hello(HelloRequest()))
    # print(await service.say_hello(message=HelloRequest()))

    # print("kek")

    # Prepare several say_hello requests
    # requests = [server.say_hello(HelloRequest(name=f"User {i}")) for i in range(1000)]

    # Await all at once
    # responses = await asyncio.gather(*requests)

    # print("lmmao")
    await h2pcontroller.close()


if __name__ == "__main__":
    add(1, 2)  # Logs arguments only
    multiply(3, 4)  # Logs return value only
    subtract(5, 2)  # Logs both arguments and return value
    asyncio.run(main())
