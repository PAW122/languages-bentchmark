# languages-bentchmark

# tester-commands:
run tests 
> tester.exe run_tests

check tests results
> tester.exe test_server_output ../bentchmarks/node-js/results.json

add new test to tests.json
> tester.exe add_task matrix_multiplication 100x100

# todo
1. add more different tasks to test potential weaknesses of languages
2. add benchmark implementation in more languages so that potentially interested people have a starting point for doing optimization
3. organize some form of testing that will be reliable and easy to 
    > docker / cloud / ?