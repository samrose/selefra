#!/usr/bin/env bash
#######################################################################################################################
#                                                                                                                     #
#                              This script helps you test interactive programs                                        #
#                                                                                                                     #
#                                                                                                                     #
#                                                                                                   Version: 0.0.1    #
#                                                                                                                     #
#######################################################################################################################

# for command `selefra logout`
cd ../../../
go build
rm -rf ./test
mkdir test
mv selefra.exe ./test
cd test
echo "begin run command selefra logout"
./selefra.exe logout $@

