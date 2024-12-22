# PBFT Consensus Optimization

This repository contains an optimized implementation of the Practical Byzantine Fault Tolerance (PBFT) consensus protocol, featuring a hash-first approach with background data transmission.

## Key Features

- Hash-first consensus optimization
- Background data transmission
- Reduced network congestion
- Improved throughput and latency
- Maintained PBFT safety guarantees

## Performance Improvements

- 28.49% increase in throughput (from 149.31 to 191.85 ops/sec)
- 22.37% reduction in mean latency (from 6.66ms to 5.17ms)
- More stable maximum latencies (from 83.09ms to 37.48ms)
- Better handling of peak loads
- Improved resource utilization

## System Requirements

- Go version 1.17 or higher
- Linux
- 8GB RAM minimum
- Network: 1Gbps Ethernet

## Installation

1. Clone the repository:
```bash
git clone https://github.com/username/PaxiBFT.git
cd PaxiBFT
```

2. Switch to the optimization branch:
```bash
git checkout optimize/background-data-transfer
```

3. Install dependencies:
```bash
go mod download
```

4. Build the project:
```bash
make build
```

## Usage

1. Start the PBFT network:
```bash
cd bin
```

2. Run performance tests:
```bash
./test.sh
```

## Architecture

The optimized protocol implements three main components:

1. Data Transmission Layer
2. State Management
3. Consensus Protocol

Key optimizations include:
- Separate hash transmission from data transfer
- Asynchronous background data propagation
- Efficient state synchronization

## Performance Testing

Test configuration:
- Concurrency Level: 1
- Write Ratio: 1.0
- Number of Keys: 1000
- Benchmark Duration: 60 seconds
- Test repetitions: Multiple runs for consistency

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request

## Authors

- Sufyan Farea Mohammed
- Faris Saad Muaddi

## Acknowledgments

Special thanks to Dr. Salem Alqahtani for supervision and guidance throughout this project.