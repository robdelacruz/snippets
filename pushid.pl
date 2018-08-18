#!/usr/bin/env perl

use v5.14;

use Time::HiRes qw/gettimeofday/;

# ms since epoch
sub epochMs {
  my ($sec, $microSec) = gettimeofday;
  return int($sec*1000 + $microSec/1000);
}

sub gen_id {
  state $_pushchars = "-0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz";
  state $_lastPushTime;
  state @_randIndexes;

  my $now = epochMs();
  if ($now != $_lastPushTime) {
    # Generate new set of random indexes to pushchars.
    for (my $i=0; $i < 12; $i++) {
      $_randIndexes[$i] = int(rand(64));
    }
  } else {
    # Same time, so just increment previous random index.
    for (my $i=0; $i < 12; $i++) {
      $_randIndexes[$i] += 1;
      if ($_randIndexes[$i] < 64) {
        last;
      }
      $_randIndexes[$i] = 0;  # carry inc to next digit
    }
  }
  $_lastPushTime = $now;

  my @id;
  # Copy pushchars to @id columns [19..8] (rightmost index to left)
  for my $idx (@_randIndexes) {
    unshift @id, substr($_pushchars, $idx, 1);
  }

  # Add time-based pushchars to @id columns [7..0] (rightmost index to left)
  for (my $i=0; $i < 8; $i++) {
    unshift @id, substr($_pushchars, $now % 64, 1);
    $now = int($now / 64);
  }

  my $sid = join("", @id);
  return $sid;
}

sub is_id {
  my ($sid) = @_;

  return $sid =~ /^[-\w]{20}$/;
}

for (my $i; $i < 10; $i++) {
  say gen_id()
}

