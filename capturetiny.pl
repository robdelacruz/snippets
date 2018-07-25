#!/usr/bin/env perl
use v5.14;

use Capture::Tiny qw(capture);

my ($out, $err, $ret) = capture {
  system("./exp-find.pl cat:food -t");
  #system("./exp-add.pl food");
};

say "out: $out";
say "err: $err";
say "ret: $ret";

