Hello World in Brainfuck (the version from the Brainfuck article on Wikipedia)

++++++++                Set cell 0 to 8
[
    >++++               Add 4 to cell 1; this will always set cell 1 to 4
    [                   as the cell will be cleared by the loop
        >++             Add 2 to cell 2
        >+++            Add 3 to cell 3
        >+++            Add 3 to cell 4
        >+              Add 1 to cell 5
        <<<<-           Decrement the loop counter in cell 1
    ]                   Loop until cell 1 is zero; number of iterations is 4
    >+                  Add 1 to cell 2
    >+                  Add 1 to cell 3
    >-                  Subtract 1 from cell 4
    >>+                 Add 1 to cell 6
    [<]                 Move back to the first zero cell you find; this will
                        be cell 1 which was cleared by the previous loop
    <-                  Decrement the loop counter in cell 0
]                       Loop until cell 0 is zero; number of iterations is 8

>>.                     Cell 2 has value 72 which is 'H'
>---.                   Subtract 3 from cell 3 to get 101 which is 'e'
+++++++..+++.           Likewise for 'llo' from cell 3
>>.                     Cell 5 is 32 for the space
<-.                     Subtract 1 from cell 4 for 87 to give a 'W'
<.                      Cell 3 was set to 'o' from the end of 'Hello'
+++.------.--------.    Cell 3 for 'rl' and 'd'
>>+.                    Add 1 to cell 5 gives us an exclamation point
>++.                    And finally a newline from cell 6
