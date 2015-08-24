# tdigest #

## Error bounds ##

Errors for computing the quantile *q* are proportional to
(*q*)(1-*q*). That means that error for extreme percentiles (like
99th, or 99.9th) are small: *q*=0.999 gives an error proportional
to 0.001*0.999 ~= 0.001


## Caveats ##
 * If the input data are sorted, or even nearly sorted, then the error
   in estimates can be unbounded if the computation is done in
   parallel (if the computation is done serially, you're safe).
