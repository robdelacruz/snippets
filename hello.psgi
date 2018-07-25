use Plack::Request;

#my $app = sub {
#  my $env = shift;
#  my $req = Plack::Request->new($env);
#
#  my $name = $req->param("name");
#  my $res = $req->new_response(200);
#  $res->content_type("text/html");
#  $res->content("<html><body>Hello $name</body></html>");
#
#  return $res->finalize();
#};

my $old = sub {
  my $env = shift;
  my $path = $env->{PATH_INFO};

  my @vals;
  for my $k (sort keys %$env) {
    my $v = $env->{$k};
    push @vals, "$k: $v";
  }
  my $senv = join("\n", @vals);

  my $req = Plack::Request->new($env);
  my @sparams;
  for my $k (sort keys %{$req->parameters}) {
    my $v = $req->parameters->{$k};
    push @sparams, "$k: $v";
  }
  my $sreqparams = join("\n", @sparams);


  my $sform = <<ROBFORM;
  <form action="/submit" method="post">
    <label>Name</label><br>
    <input name="name" type="text"><br>
    <label>Address</label><br>
    <input name="address" type="text"><br>
    <button>OK</button>
  </form>
ROBFORM

  if ($path eq "/") {
    return [200, ["Content-Type" => "text/html"], [$sform]];
  } elsif ($path eq "/abc") {
    return [200, ["Content-Type" => "text/plain"], ["abc"]];
  } else {
    return [202, ["Content-Type", "text/plain"], [$sreqparams, $senv]];
  }

};

