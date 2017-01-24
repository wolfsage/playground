use strict;
use warnings;

my $name = 'privmsg-emote-on-slack.pl';
my $VERSION = '0.1';

# Inspired by rjbs' irssi script:
#
# https://rjbs.manxome.org/rubric/entry/2110
#
# Slack eats emote (/me ...) in privmsg to users for some reason.
# If you try to emote on a slack irc gateway, instead send
# it as a normal message wrapped in '*'s.
#
# For example:
#
#   /me goes dancing
#
# will get sent as
#
#  *goes dancing*

weechat::register(
  $name,
  'Matthew Horsfall (alh) <WolfSage@gmail.com>',
  $VERSION,
  'GPL3',
  'Make sure emotes in privmsgs in Slack servers make it',
  '', '',
);

weechat::hook_command_run("/me", "ensure_emote_makes_it", "");

sub ensure_emote_makes_it {
  my (undef, $buffer, $arg) = @_;

  # Kill command
  $arg =~ s/^[^\s]+\s//;

  my $type = weechat::buffer_get_string($buffer, "localvar_type");

  # Emotes only get eaten on private messages
  return weechat::WEECHAT_RC_OK unless $type eq 'private';

  my $server = weechat::buffer_get_string($buffer, "localvar_server");
  my $infolist = weechat::infolist_get( "irc_server", "", "$server" );
  weechat::infolist_next($infolist);

  my $addresses = weechat::infolist_string($infolist, "addresses");

  return weechat::WEECHAT_RC_OK unless $addresses && $addresses =~ /irc\.slack\.com/i;

  # Send it normally
  weechat::command($buffer, "/msg * *$arg*");

  # And don't try to send the emote
  return weechat::WEECHAT_RC_OK_EAT;
}
