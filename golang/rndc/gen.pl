#!/usr/bin/perl

use strict;
use warnings;

use Net::RNDC::Packet;

my $pkt = Net::RNDC::Packet->new(key => 'abcd');

$pkt->{data}{somelist} = [ {cat => 'dog', small => ['thing'], }, 1, "hi", ['bird', 'mouse'] ];

open(my $o, '>', 'pkt.pkt') or die "Failed to write pkt.pkt: $!\n";

print $o $pkt->data;

close($o);

print <<EOF
Now run:

  go build
  ./parse
EOF
