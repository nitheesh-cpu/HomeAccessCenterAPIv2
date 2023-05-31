# HomeaccessCenterAPIv2

HomeaccessCenterAPIv2 is an improved API that scrapes HomeaccessCenter for data. It is written in Go and offers faster performance compared to v1.

## Documentation

The API documentation is available at [https://homeaccesscenterapi-docs.vercel.app/](https://homeaccesscenterapi-docs.vercel.app/). It provides detailed information about the available endpoints, request/response formats, and usage examples.

## API Endpoint

The API can be accessed at [https://homeaccesscenterapi.vercel.app/api/](https://homeaccesscenterapi.vercel.app/api/). Please refer to the documentation for the available endpoints and their purposes.

## Performance Comparison

The following table compares the performance of Python and Go for each endpoint:

| Endpoint       | Python (v1) | Go (v2) |
|----------------|-------------|---------|
| /info          | 1800 ms     | 1500 ms |
| /rank          | 2000 ms     | 1500 ms |
| /transcript    | 2100 ms     | 1500 ms |
| /name          | 2400 ms     | 1500 ms |
| /ipr           | 2600 ms     | 2000 ms |
| /reportcard    | 3200 ms     | 2700 ms |
| /averages      | 3800 ms     | 3300 ms |
| /classes       | 3900 ms     | 3300 ms |
| /assignments   | 3900 ms     | 3300 ms |


The measurements represent the average response time for each endpoint. As you can see, the Go version (v2) of the API outperforms the Python version (v1) in terms of response time. It provides faster and more efficient performance, allowing for quicker retrieval of grades, attendance, assignments, and schedule information.
