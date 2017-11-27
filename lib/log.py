#!/usr/bin/env python3

import inspect
import os
import time
from functools import partial


def _log(level, message):
    """log
    :param level: 消息级别，info、warn、error
    :type level: str
    :param message: 消息内容
    :type message: str
    """
    filename, line_number, function_name, *_ = inspect.getframeinfo(inspect.currentframe().f_back)
    print("[{}] [{}] [{}:{}:{}] {}".format(time.strftime("%Y-%m-%d %H:%M:%S"), level, os.path.basename(filename),
                                           function_name, line_number, message))


info = partial(_log, "info")
warn = partial(_log, "warn")
error = partial(_log, "error")
