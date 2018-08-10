#!/usr/bin/env tclsh

set cmd {|grep rob}

# Open pipe for read and write.
set f [open $cmd r+]

# Write some lines to pipe.
puts $f {
  Now is the time
  for all good men with rob
  to come to the aid of the party.
  rob 12345
}
flush $f

# Close channel to write EOF and prevent read from blocking.
chan close $f write

set output [read $f]
puts $output

close $f

# 1. Run the command exp-find.pl
# 2. err is 1 if error encountered. stderr output is set to cmdout.
# 3. If no error, cmd stdout is set to cmdout.
set err [catch {exec ../expbud/exp-find.pl --ytd -t} cmdout]
puts "err: $err"
puts "cmdout: $cmdout"

