use strict;
use warnings;

use charnames ':full';

my $name = 'emoji.pl';
our $VERSION = '0.1';

# Adapted from rjbs' irssi script:
#
# https://github.com/rjbs/rjbs-dots/blob/master/.irssi/scripts/slack-emoji.pl
#
# Changes incoming slack emoji names to unicode, (:+1:, :smile:, etc...)
#
# Your server name must have the string 'slack' in it to work!

my %emoji = (
  'imp'   => "\N{IMP}",
  'heart' => "\N{BLUE HEART}", # Pobox, yo.
  'poop'  => "\N{PILE OF POO}",
  'smile' => "\N{SMILING FACE WITH OPEN MOUTH AND SMILING EYES}",
  '+1'    => "\N{THUMBS UP SIGN}",
  '-1'    => "\N{THUMBS DOWN SIGN}",
  'snowman' => "\N{SNOWMAN}",
);

weechat::register(
  $name,
  'Matthew Horsfall (alh) <WolfSage@gmail.com>',
  $VERSION,
  'GPL3',
  'Convert Slack Emoji to Unicode',
  '', '',
);

weechat::hook_modifier("irc_in_privmsg", "change_emoji", "");

sub change_emoji {
  my ($data, $modifier, $modifier_data, $string) = @_;

  # Only do this for slack servers
  return $string unless $modifier_data =~ /slack/;

  return munge_emoji($string);
}

sub munge_emoji {
  my ($target, $text) = split / :/, $_[0], 2;
  $text =~ s!:([-+a-z0-9_]+):!$emoji{$1} // ":$1:"!ge;
  return "$target :$text";
}
