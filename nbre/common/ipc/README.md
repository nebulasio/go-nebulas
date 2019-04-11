# What's shared memory for?
Shared memory is for interprocess communication, like exchanging data or
commands. Specially, shared memory is for multiple processes on the same
machine.

In Nebulas, we follow the thumb role in system design, *modularity*, which
means we separate difference functionalities into different processes. Thus, we
shall have multiple processes even on one node (physical machine). Given this,
interprocess communication is quite critical to Nebulas mainnet.

# Why shared memory?
As we stated, interprocess communication (IPC) is critical to Nebulas mainnet.
However, there are different ways for IPC, and some of the most popular way is
to use networking, like RPC, or gRPC from Google. Then why shared memory?

The answer is *performance*. Most networking libraries, like RPC and gRPC are
designed for cross-machine communication, which means they need to meet design
goals that totally different from locale-machine IPC. For example, networking needs to
do data serialization and deserialization, consider
  different endians (big-endian vs. little endian), etc. What's worse, they
  may need to conform standard protocols (like HTTP), and common security
  standards (like SSL). Even we can ignore some limitations given
  locale-machine, we still need copy data to the networking buffer on the sender
  side, and copy data from the networking buffer on the client side. This could
  cause performance penalty.

To the best of our knowledge, shared memory is the best IPC practice when
pursuing performance.

# The design space of Nebulas shared memory communication
The two main goals of designing this shared memory communication mechanism are
performance and stability. To achieve these goals, we simplify the model of
shared memory communication.

In our communication model, we have one *server* and one *client*. The server
side is responsible to initiate/close the enviornments. Both server and client
can write and read data from the shared memory.

Performance is relatively easy to archieve. However, there are concerns since
we are using Boost, instead of low-level POSIX APIs. We may need revise this in
the future.

For stability, there are tradeoffs to make. The key is how we handle failures
of interacting processes. For example, what should we do when the client/server
crash? Notice that it could be different scenario for the client crash and the
server crash. For client crash, the server could restart the client without
initiating the handlers. Yet, for the server crash, should we restart the
server directly, or should we restart both the client and the server? Here we
choose restarting both the client and the server. There are two reasons. First,
the server may fail to initiate the handlers since the client still hold them.
Second, the server crash means the whole functionalities stop working, and it's
meaningless to keep the client running.

Thus, it is important to be aware of when the other side is crashed or not.
An intuitive way is to involve heart beating in typical distributed
systems. We use the same idea here, yet with different implementation.

Consequently, we have two design choices. First, the server cannot start when there is already
another server instance or any client instance. Second, both the server and the client may raise exceptions when the other side crash (heart beat timeout). And the server may choose restart the client and the client may choose crash immediately.

# Future work
We need a comprehensive performance evaluation compared to networking and
native implementation.
