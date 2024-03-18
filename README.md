<h2>LayDB is a lightweight in-memory key-value database developed in the Go programming language.<br /> </h2>
It is designed to provide fast and efficient storage and retrieval of key-value pairs entirely in memory, catering to the needs of modern, high-performance applications.

<h3>Key Features:</h3><br /> 
1. High Performance: LayDB offers ultra-fast read and write operations by storing all data in memory, eliminating the need for disk I/O.<br /> 
2. Concurrency: Leveraging Go's powerful concurrency model, LayDB handles multiple concurrent read and write requests with ease, ensuring optimal performance under heavy loads.<br /> 
3. Scalability: LayDB is designed to scale effortlessly with the growth of your application, accommodating large volumes of data and high throughput.<br /> 
4. Persistence : LayDB provides an optional persistence layer, allowing data to be saved to disk for durability, ensuring data integrity even in the event of system failure.<br /> 
5. Simple API: LayDB offers a simple and intuitive API for storing, retrieving, and deleting key-value pairs, making it easy to integrate with existing applications.<br />

<h3>Project Goals:</h3><br />
1. Develop a robust and performant in-memory key-value database entirely in Go.<br /> 
2. Optimize memory usage and performance to ensure efficient operation, even under heavy workloads.<br /> 
3. Implement comprehensive testing to validate functionality and reliability, ensuring a stable and dependable database.<br /> 
4. Replace native Hash map with B+ tree implementation.<br /> 
