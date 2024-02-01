
# coroutine to send query results to  MetricsQuery channel
import asyncio
import aiohttp
from collections import defaultdict, namedtuple
from typing import Any, Dict,Callable

QueryResult = namedtuple('QueryResult', ['query', 'result', 'err'])
# Channel to send query results
class MetricsQuery:
    # Initialize the channel
    def __init__(self, callback: Callable[[QueryResult], None], headers: Dict[str, str]) -> None:
        self.callback = callback
        self.queue = asyncio.Queue()
        self.headers = headers
    # Send query result to the channel
    async def send(self, result: QueryResult) -> None:
        await self.queue.put(result)
    # Close the channel
    async def close(self):
        await self.queue.put(None) # No more items to iterate over
        
    # Receive query result from the channel
    def __aiter__(self):
        return self
    async def __anext__(self):
        result = await self.queue.get()
        if result is None:  # No more items to iterate over
            raise StopAsyncIteration
        return result
    # Query prometheus
    async def Prometheus_Query(self, session: aiohttp.ClientSession, query: str) -> None:
        try:
            result = await self.api_return(session, query)
            await self.send(QueryResult(query, result, None))
        except Exception as e:
            await self.send(QueryResult(query, None, e))
    # return the query from prometheus api
    async def api_return(self, session, url):
        async with session.get(url, headers=self.headers, verify_ssl=False) as resp:
            if resp.status != 200:
                raise Exception(f"Error occurred while querying {url}: {resp.status}")
            return await resp.json()

# Process the query result
def processResult(result: Any) -> None:
    if result.err:
        print(f"Error occurred while querying {result.query}: {result.err}")
        return
    if result.result['status'] == 'success':
        subnet_cidr, pod_ip_count = processVectorResult(result.result['data']['result'])
        print(f"subnet_cidr per subnet: {subnet_cidr}")
        print(f"pod_ip_count per subnet: {pod_ip_count}")
    else:
        #print(f"Error occurred while querying {result.query}: {result.result}")
        print(f"Error occurred while querying {result.query}: {result.result['errorType']}: {result.result['error']}")

# Process the vector result from prometheus
def processVectorResult(vector):
    subnet_cidr = defaultdict(str)
    pod_ip_count = defaultdict(int)
    
    # Iterate over the vector and store the subnet_cidr and pod_ip_count
    for elem in vector:
        if 'subnet' in elem['metric'].keys():
            subnet=elem['metric']['subnet']
            subnet_cidr[subnet] = elem['metric']['subnet_cidr']
            pod_ip_count[subnet] += int(elem['value'][1])
        elif elem['metric']['subnet_cidr']:
            subnet_cidr[elem['metric']['subnet_cidr']] = elem['value'][1]
            pod_ip_count[elem['metric']['subnet_cidr']] += int(elem['value'][1])
        else:
            print(f"Error occurred while querying {elem}: {elem['metric']}")
    return  subnet_cidr, pod_ip_count
        
# Main function
async def main() -> None:
    headers = {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer <token>'
    }
    base_url = "http://127.0.0.1:9090/api/v1/query"
    # query as list input
    queries =[f"{base_url}?query=cx_ipam_total_ips{{subnet_cidr=~\".*\"}}",
            f"{base_url}?query=sum(cx_ipam_total_ips{{subnet_cidr=~\".*\"}}) by (subnet_cidr)"]
    ch = MetricsQuery(print, headers)
    async with aiohttp.ClientSession() as session:
        try:
            # Use asyncio.gather to send multiple queries in parallel
            await asyncio.gather(*(ch.Prometheus_Query(session, query) for query in queries))
            await ch.close()
            async for result in ch:
                processResult(result)
        except Exception as e:
            print(f"Error occurred while querying {queries}: {e}")

if __name__ == "__main__":
    asyncio.run(main())