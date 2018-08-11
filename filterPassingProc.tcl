#!/usr/bin/env tclsh

# Doing filter on a list, passing a filter proc.
# Thanks to @chatrwing for the tip on getting this to work.

proc filter {l filterCmd} {
  set out [list]

  foreach n $l {
    if {[eval $filterCmd n]} {
      lappend out $n
    }
  }

  return $out
}

proc match {q nVar} {
  # Need to pass element by reference because match will be called
  # by eval or uplevel in filter function.
  upvar $nVar n

  # Put filter code here. This simple example searches strings
  # but usually you'll search a struct's (array/dict) elements.
  if {[regexp -nocase -- $q $n]} {
    return 1
  }
  return 0
}

# Usually the list will be a more complicated struct, such as
# a list of arrays, or list of dicts.
set alist {abc def ghi add ddd}

set q "dd"
if {[llength $argv] > 0} {
  set q [lindex $argv 0]
}
set out [filter $alist "match $q"]

puts "Input list: $alist"
puts "Filtered to include '$q'"
puts "Output list: $out"

