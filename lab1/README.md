# 🌐 Self-Copy Network Detector
## 🚀 Overview
Self-Copy Network Detector is a powerful network application that automatically discovers and monitors other running instances of itself within your local network. Using efficient multicast UDP communication, it maintains real-time awareness of all active copies and provides instant notifications when instances appear or disappear from the network.

## ⚡ How It Works
- 🎯 **Multicast Group Joining - The application joins a specified multicast group and begins exchanging heartbeat messages with other instances

- 📡 **Real-time Heartbeats - Each instance periodically sends and listens for heartbeat messages to maintain network awareness

- 🔄 **Dynamic Registry - Every node maintains a live registry of participating instances and dynamically updates as the network changes

- 🌍 **Dual Protocol Support - Automatically adapts to IPv4 or IPv6 networks based on the provided multicast address

## Usage
go run main.go <address multicast group> <member port>
Simply launch the application with a multicast group address parameter and member multicast group port, and it will begin monitoring for other instances. The application immediately starts displaying detected copies and continues to provide real-time updates as the network environment changes.
