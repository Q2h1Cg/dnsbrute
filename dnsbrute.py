#!/usr/bin/env python3

import asyncio

from lib import dns


loop = asyncio.get_event_loop()
client = dns.Client()
client2 = dns.Client()
loop.run_until_complete(client._query_ns("sh3ll.me"))
print(client._resolvers)
print(client2._resolvers)

