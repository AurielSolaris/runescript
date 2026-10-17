Print the alphabet

Comments in Brainfuck are any characters that are not commands so these
avoid the eight command characters

++++++++[>++++++++<-]>+     cell 1 = 8 times 8 plus 1 = 65 = the letter A
>+++++[>+++++<-]>+          cell 3 = 5 times 5 plus 1 = 26 = the count
[<<.+>>-]                   print the letter and step to the next one
                            then count down until the count is zero
++++++++++.                 cell 3 = 10 = a newline
