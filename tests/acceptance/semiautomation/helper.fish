function printheader
  echo "Test            : $TESTNAME"
  echo "Description     : $TESTDESC"
  echo "Yaml file       : $YAMLFILE"
  echo "Deployment name : $DEPLOYMENT"
  echo
end

function waitForKubectl
  if test (count $argv) -lt 5
    return 1
  end
  set -l op (string split -- " " $argv[1])
  set -l select $argv[2]
  set -l good (string split -- ";" "$argv[3]")
  set -l expected $argv[4]
  set -l timeout $argv[5]
   
  echo
  echo "Testing `kubectl $op`"
  echo "  for occurrences of `$select`"
  echo "  that are `$good`, expecting `$expected`"
  echo 

  set -l t 0
  while true
    set -l l (kubectl $op | grep $select)
    set -l nfound (count $l)
    set -l ngood 0
    for line in $l
      if string match -r $good $line > /dev/null
        set ngood (math $ngood + 1)
      end
    end
    echo -n "Good=$ngood, found=$nfound, expected=$expected, try $t ($timeout)"
    echo -n -e "\r"
    if test $ngood -eq $expected -a $nfound -eq $expected ; echo ; return 0 ; end
    if test $t -gt $timeout ; echo ; echo Timeout ; return 2 ; end
    set t (math $t + 1)
    sleep 1
  end
end

function output
  if which say > /dev/null
    say $argv[1] > /dev/null ^ /dev/null
  end
  for l in $argv[2..-1] ; echo $l ; end
end

function log
  echo "$argv[1] Test: $TESTNAME, Desc: $TESTDESC" >> testprotocol.log
end

function inputAndLogResult
  read -P "Test result: " result
  log $result
end

function waitForUser
  read -P "Hit enter to continue"
end

function getLoadBalancerIP
  string trim -c '"' (kubectl get service $argv[1] -o=json | \
                      jq .status.loadBalancer.ingress[0].ip)
end

function testArangoDB
  set -l ip $argv[1]
  set -l timeout $argv[2]
  set -l n 0
  echo Waiting for ArangoDB to be ready...
  while true
    if set v (curl -k -s -m 3 "https://$ip:8529/_api/version" --user root: | jq .server)
      if test "$v" = '"arango"' ; return 0 ; end
    end
    set n (math $n + 1)
    if test "$n" -gt "$timeout"
      echo Timeout
      return 1
    end
    echo Waiting "$n($timeout)"...
  end
end

function fail
  output "Failed" $argv
end
