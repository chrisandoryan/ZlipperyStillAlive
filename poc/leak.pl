    use strict;
    use warnings;
     
    my $filename = 'Flag';
    if (open(my $fh, '<:encoding(UTF-8)', $filename)) {
      while (my $row = <$fh>) {
        chomp $row;
        print "$row\n";
      }
    } else {
      warn "Could not open file '$filename' $!";
    }
