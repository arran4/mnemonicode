cmake_minimum_required(VERSION 3.5)
project(mnemonic C)

set(CMAKE_C_STANDARD 90)
set(CMAKE_C_STANDARD_REQUIRED ON)
set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -Wall -Wextra -pedantic")

add_library(mnemonic_lib STATIC mnemonic.c mn_wordlist.c)

add_executable(mnencode mnencode.c)
target_link_libraries(mnencode mnemonic_lib)

add_executable(mndecode mndecode.c)
target_link_libraries(mndecode mnemonic_lib)
