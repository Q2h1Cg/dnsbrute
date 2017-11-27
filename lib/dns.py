#!/usr/bin/env python3

import asyncio
import threading

import aiodns
import async_timeout
from pycares import QUERY_TYPE_NS, QUERY_TYPE_A, QUERY_TYPE_CNAME
from publicsuffix import PublicSuffixList

from lib import log


_resolver = None
_black_list = {}
_psl = None


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


async def query_loop(domain, queue):
    """轮询
    :param domain: 域名
    :type domain: str
    :param queue: 子域名队列
    :type queue: asyncio.Queue
    """
    if not _resolver:
        thread = threading.Thread(target=_query_ns, args=(domain, asyncio.get_event_loop()))
        thread.start()
        thread.join()

    # 循环读取 queue
    while True:
        sub_domain = await queue.get()
        if sub_domain is None:
            return

        record = await query_a_cname(sub_domain)
        if record:
            log.info(record)
        # queue.task_done()


def _query_ns(domain, loop):
    """查询域名 NS 记录
    :param domain: 域名
    :type domain: str
    """
    loop_ = asyncio.new_event_loop()
    asyncio.set_event_loop(loop_)
    ns_records = []
    ns_servers = []
    exception = None
    for _ in range(3):
        try:
            ns_records = loop_.run_until_complete(aiodns.DNSResolver().query(domain, "NS"))
        except Exception as ex:
            exception = ex
        else:
            break

    # 读取 ns 记录时出现异常或无 NS 记录
    if exception or not ns_records:
        log.error("{}, {}, {}".format(domain, ns_records, exception))
        return

    for ns_record in ns_records:
        a_records = loop_.run_until_complete(aiodns.DNSResolver().query(ns_record.host, "A"))
        for a_record in a_records:
            ns_servers.append(a_record.host)

    log.debug("ns servers for {}: {}".format(domain, ns_servers))
    global _resolver
    _resolver = aiodns.DNSResolver(ns_servers, loop)


async def query_a_cname(domain):
    """查询 DNS A、CNAME 记录
    :param domain: 域名
    :type domain: str

    :return: 查询结果
    :rtype Record
    """
    parent_domain = _parent_domain(domain)
    if parent_domain not in _black_list:
        # 添加黑名单
        _query_pan_dns(parent_domain)

    record = None
    for query_type in ("A", "CNAME"):
        try:
            with async_timeout.timeout(5):
                records = await _resolver.query(domain, query_type)
        except:
            pass
        else:
            if query_type == "CNAME":
                record = Record(domain, QUERY_TYPE_CNAME, records.ttl, records.cname)
            elif query_type == "A" and records:
                record = Record(domain, QUERY_TYPE_A, records[0].ttl, [record_.host for record_ in records])
            break

    return record


def _parent_domain(domain):
    """域名的父域名
    :param domain: 域名
    :type domain: str

    :return: 父域名
    :rtype: str
    """
    if _is_root_domain(domain):
        return domain

    return domain[domain.index(".")+1:]


def _is_root_domain(domain):
    """是否是主域名
    :param domain: 域名
    :type domain: str

    :return: 是否是主域名
    :rtype: bool
    """
    global _psl
    if not _psl:
        with open("dict/public_suffix_list.dat") as fd:
            _psl = PublicSuffixList(fd)

    return domain == _psl.get_public_suffix(domain)


def _query_pan_dns(domain):
    """生成泛解析黑名单
    :param domain: 域名
    :type domain: str
    """
    global _black_list
    _black_list[domain] = {}


def _is_pan_dns(record):
    """判断是否是泛解析
    :param record: 域名记录
    :type record: Record

    :return: 是否是泛解析
    :rtype: bool
        """
    return False
