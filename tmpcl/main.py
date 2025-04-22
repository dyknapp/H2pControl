import asyncio

# h2pcontrol python library to connect to manager and handle other server dependencies
from h2pcontrol_connector import H2PControl

# The library you want to use, here we use a helloworld example
from hpcontrol_python_server.arduino import GreeterStub, HelloRequest

async def main():
    # First connect to the H2P manager
    h2pcontroller = H2PControl("localhost:50051")
    await h2pcontroller.connect()
    
    
    # Register any channel/service combination with full type safety by passing its stub
    # Find a list of servers with `h2pcontrol fetch`. To get more specific information of a service, use
    # `h2pcontrol fetch <name>`
    channel, service = await h2pcontroller.register_server("arduino", GreeterStub)
    
    # GRPC async calls, so use await if you need the response.
    print(await service.say_hello(HelloRequest()))

    # To run requests parallel, you can just simply 'gather' requests and await at once:
    # Prepare several say_hello requests
    requests = [service.say_hello(HelloRequest(name=f"User {i}")) for i in range(3)]
    
    # Await all at once
    responses = await asyncio.gather(*requests)
    
    for response in responses:
        print(response)
        
    # In the case you just want to fire-and-forget a function call, we can create an asyncio task:
    asyncio.create_task(service.say_hello(HelloRequest(name="Background task")))
  
    await h2pcontroller.close()


if __name__ == "__main__":
    asyncio.run(main())


