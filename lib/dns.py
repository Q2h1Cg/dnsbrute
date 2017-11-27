#!/usr/bin/env python3

import aiodns

from pycares import QUERY_TYPE_NS, QUERY_TYPE_A, QUERY_TYPE_CNAME


class Record:
    """DNS 记录"""

    def __init__(self, domain, query_type, ttl, answer):
        """
        :param domain: 域名
        :type domain: str
        :param query_type: 记录类型
        :type query_type: int
        :param ttl: 生存时间
        :type ttl: int
        :param answer: 请求结果
        :type answer: list
        """
        self.domain = domain

        if query_type not in (QUERY_TYPE_NS, QUERY_TYPE_A, QUERY_TYPE_CNAME):
            raise ValueError('query_type must be one of QUERY_TYPE_NS, QUERY_TYPE_A or QUERY_TYPE_CNAME')

        self.type = query_type
        self.ttl = ttl
        self.answer = answer if answer else []

    def __str__(self):
        query_type = {QUERY_TYPE_NS: "NS", QUERY_TYPE_A: "A", QUERY_TYPE_CNAME: "CNAME"}
        return "{} - {} - {} - {}".format(self.domain, query_type[self.type], self.ttl, self.answer)

    def __repr__(self):
        return "<{}>".format(self.__str__())


async def query_ns(domain):
    """查询域名 NS 记录
    :param domain: 域名
    :type domain: str

    :return: NS 记录
    :rtype: list
    """
    records = []

    query_result = await aiodns.DNSResolver().query(domain, "NS")
    for record in query_result:
        records.append(Record(domain, QUERY_TYPE_NS, 0, record.host))

    return records
