#!/usr/bin/env python3

import asyncio

import aiodns

from lib import dns


async def produce(queue):
    # for i in range(10000):
    with open("dict/53683.txt") as fd:
        for line in fd:
            line = line.strip()
            domain ="{}.baidu.com".format(line)
            await queue.put(domain)
        # print("put {}".format(domain))
    for i in range(1000):
        await queue.put(None)


queue = asyncio.Queue()
tasks = [dns.query_loop("baidu.com", queue) for _ in range(1000)]
tasks.append(produce(queue))
loop = asyncio.get_event_loop()
loop.run_until_complete(asyncio.wait(tasks))
loop.close()

