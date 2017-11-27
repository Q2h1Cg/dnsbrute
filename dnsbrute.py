#!/usr/bin/env python3

import asyncio

from lib import dns


loop = asyncio.get_event_loop()
result = loop.run_until_complete(dns.query_ns("sh3ll.me"))
print(result)
