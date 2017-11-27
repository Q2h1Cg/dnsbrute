#!/usr/bin/env python3

import asyncio

import aiodns
from pycares import QUERY_TYPE_NS, QUERY_TYPE_A, QUERY_TYPE_CNAME

from lib import log


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

    _resolver = None
    _black_list = {}

    def __init(self, ns_servers):
        """
        :param ns_servers: NS 服务器
        :type ns_servers: list(str)
        """
        self._resolver = aiodns.DNSResolver(ns_servers)

    def _is_root_domain(self, domain):
        """是否是主域名
        :param domain: 域名
        :type domain: str

        :return: 是否是主域名
        :rtype: bool
        """
        return False

    def _parent_domain(self, domain):
        """域名的父域名
        :param domain: 域名
        :type domain: str

        :return: 父域名
        :rtype: str
        """
        if self._is_root_domain(domain):
            return domain

        return domain[domain.index(".")+1:]

    def _query_pan_dns(self, domain):
        """生成泛解析黑名单
        :param domain: 域名
        :type domain: str
        """
        self._black_list[domain] = {}

    def _is_pan_dns(self, record):
        """判断是否是泛解析
        :param record: 域名记录
        :type record: Record

        :return: 是否是泛解析
        :rtype: bool
        """
        pass

    async def query_a_cname(self, domain):
        """查询 DNS A、CNAME 记录
        :param domain: 域名
        :type domain: str

        :return: 查询结果
        :rtype list(Record)
        """
        parent_domain = self._parent_domain(domain)
        if parent_domain not in self._black_list:
            # 添加黑名单
            self._query_pan_dns(parent_domain)

        records = []
        for query_type in ("A", "CNAME"):
            try:
                _records = await self._resolver.query(domain, query_type)
            except:
                pass
            else:
                records = [Record(domain, {"A": QUERY_TYPE_A, "CNAME": QUERY_TYPE_CNAME}[query_type], record.ttl,
                                  record.host) for record in _records]
                break

        # 判断是否是泛解析
        return [record for record in records if not self._is_pan_dns(record)]


def query_ns(domain):
    """查询域名 NS 记录
    :param domain: 域名
    :type domain: str
    """
    loop = asyncio.get_event_loop()
    ns_records = []
    ns_servers = []
    exception = None
    for _ in range(3):
        try:
            ns_records = loop.run_until_complete(aiodns.DNSResolver().query(domain, "NS"))
        except Exception as ex:
            exception = ex
        else:
            break

    # 读取 ns 记录时出现异常或无 NS 记录
    if exception or not ns_records:
        log.error("{}, {}, {}".format(domain, ns_records, exception))
        return

    for ns_record in ns_records:
        a_records = loop.run_until_complete(aiodns.DNSResolver().query(ns_record.host, "A"))
        for a_record in a_records:
            ns_servers.append(a_record.host)

    log.debug("ns servers for {}: {}".format(domain, ns_servers))
    return ns_servers
