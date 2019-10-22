#!/bin/bash

gnuplot <<EOF
    set xrange [0:44100]
    set size ratio 0.5
    set terminal png
    set output 'wavedata.png'
    plot "wavedata" with lines
EOF
