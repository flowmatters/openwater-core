#!/bin/bash

set -e

source .ow-test/bin/activate
OW_BIN=${PWD} python -c 'from openwater.discovery import discover; print(discover())'

OW_BIN=${PWD} python -m openwater.tests.system_test