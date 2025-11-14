
# strawk: STRuctured AWK

## Demo 

strawk now has a live demo! [Try it out here](https://ahalbert.github.io/strawk/demo/strawk.html).

## About 

An AWK implementation using Structural Regular Expressions rather than processing things line-by-line

Rob Pike wrote a paper [Structural Regular Expressions](https://doc.cat-v.org/bell_labs/structural_regexps/se.pdf) that criticized the Unix toolset for being excessively line oriented. Tools like awk and grep assume a regular record structure usually denoted by newlines. Unix pipes just stream the file from one command to another, and imposing the newline structure limits the power of the Unix shell. 

In the paper, Mr. Pike proposed an awk of the future that used structural regular expressions to parse input instead of line by line processing. As far as I know, it was never implemented. 

So I made a prototype awk language that uses structural regular expressions. At this point it only can parse and run the examples in the paper, but I plan to make it a full awk over time.

## Contact 

Feedback is always appreciated, you can contact me at armand (dot) halbert (at) gmail.com
