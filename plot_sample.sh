#!/bin/bash

gnuplot <<EOF
    set xrange [0:220]
    set terminal png
    set output 'wavedata.png'
    plot "wavedata" with linespoints
EOF
