update jobs_script set label="cmd" where is_builtin='true' and project='oes' and label="control" and name in ('thaw.sh','reset.sh','nakedstart.sh','load.sh','open.sh','close.sh','shutoff.sh','stop.sh');

update jobs_script set label="cmd" where is_builtin='true' and project='mds' and label="control" and name in ('reset.sh','nakedstart.sh','load.sh','start.sh','stop.sh','shutoff.sh');
