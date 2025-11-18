
# strawk: STRuctured AWK

## Demo 

strawk now has a live demo! [Try it out here](https://ahalbert.github.io/strawk/demo/strawk.html).

## About 

An AWK implementation using Structural Regular Expressions rather than processing things line-by-line.

Rob Pike wrote a paper, [Structural Regular Expressions](https://doc.cat-v.org/bell_labs/structural_regexps/se.pdf), that criticized the Unix toolset for being excessively line oriented. Tools like awk and grep assume a regular record structure usually denoted by newlines. Unix pipes just stream the file from one command to another, and imposing the newline structure limits the power of the Unix shell. 

In the paper, Mr. Pike proposed an awk of the future that used structural regular expressions to parse input instead of line by line processing. As far as I know, it was never implemented. 

So I made a prototype awk language that uses structural regular expressions. It attempts to imitate standard AWK, but there may be differences since it's using go libraries under the hood.

## Installing

1. Clone the repository.
2. Run `make install`. This will install the binary in your $GOBIN folder or ~/go/bin if it's not installed. You will have to add ~/go/bin to your path if it's not already there.

## Using 

Strawk is given: 
1. *A program*, a series of regular expressions and code blocks, 
2. *Input*, either a file or input from stdin.

It attempts to match the patterns given in the program against the input. Once a match is found, it expands the match until it no longer matches, then runs the code block.

```awk

# Prints at the begin
BEGIN {
    print "--header--"
}

# The first and highest priority expression - other blocks will not run if this matches.
/foobar/ {
    # do something
}

# The second and lower priority expression
/foo/ {
    # do something else
}

END {
    print "--footer--"
}
```

## Developing 

1. Compile with `make build` 
2. Run tests with `make tests`

## Contact 

Feedback is always appreciated, you can contact me at armand (dot) halbert (at) gmail.com
