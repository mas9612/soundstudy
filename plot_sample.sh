#!/bin/bash

gnuplot <<EOF
    set xrange [0:44100]
    set size ratio 0.5
    set terminal png
    set output 'wavedata_all.png'
    plot "wavedata" with lines
EOF

gnuplot <<EOF
    set xrange [0:1600]
    set size ratio 0.5
    set terminal png
    set output 'wavedata_partial.png'
    plot "wavedata" with lines
EOF
