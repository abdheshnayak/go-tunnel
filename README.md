# GoTunnel

GoTunnel is a lightweight Go program that allows you to expose a local WebSocket or TCP server to the public through a server hosted on Kubernetes. It establishes a secure tunnel between a client running locally and a server in Kubernetes, enabling external access to services that are otherwise only accessible locally.

## Features

- **Bidirectional Tunneling:** Allows data to flow between the local server and external clients seamlessly.
- **Lightweight and Simple:** Minimal dependencies and straightforward setup.
- **Persistent Connection:** Handles multiple requests without hanging or resource exhaustion.
- **Secure Communication:** Supports secure tunneling with encryption (e.g., via TLS, to be implemented).

## Getting Started

### Prerequisites

- Go 1.16 or higher
- A Kubernetes cluster with access to deploy and expose services
- A local server (e.g., WebSocket or TCP server) running on your machine

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/GoTunnel.git
   cd GoTunnel
