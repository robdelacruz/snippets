#!/usr/bin/perl
use v5.14;

use DBI;

my $dbfile = "test.db";
my $dbh = DBI->connect("dbi:SQLite:dbname=$dbfile", "", "", {RaiseError => 1})
  or die $DBI::errstr;

my $stmt = "INSERT INTO expense (id, catid, amt, body) VALUES ('1', 'commute', 2.00, 'grab')";
$dbh->do($stmt) or die $DBI::errstr;

my $query = "SELECT id, catid, amt, body FROM expense ORDER by id";
my $sth = $dbh->prepare($query);
my $err = $sth->execute() or die $DBI::errstr;

while(my @row = $sth->fetchrow_array()) {
  my ($id, $catid, $amt, $body) = @row;
  my $samt = sprintf("%.2f", $amt);

  say "$id";
  say "$catid  $samt";
  say "$body";
}

$dbh->disconnect();


