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


class Client:
    """DNS 客户端"""
    _resolvers = {}

    async def _query_ns(self, domain):
        """查询域名 NS 记录
        :param domain: 域名
        :type domain: str
        """
        if domain in self._resolvers:
            return

        ns_servers = []
        ns_records = await aiodns.DNSResolver().query(domain, "NS")
        for ns_record in ns_records:
            a_records = await aiodns.DNSResolver().query(ns_record.host, "A")
            for a_record in a_records:
                ns_servers.append(a_record.host)

        self._resolvers[domain] = aiodns.DNSResolver(ns_servers)

    async def query_a_cname(self, domain):
        """查询 DNS A、CNAME 记录
        :param domain: 域名
        :type domain: str
        """
        pass
