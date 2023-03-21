#!/bin/bash
ps -ef | grep postgre | awk '{print $2}' | xargs -i kill {}