Self-Copy Detection in Local Network

Develop an application that detects copies of itself in the local network by exchanging multicast UDP messages. The application must track the appearance and disappearance of other instances in the local network and display the list of IP addresses of "alive" copies when changes occur.

The multicast group address must be passed as a parameter to the application. The application must support both IPv4 and IPv6 networks, automatically selecting the protocol based on the provided group address.
