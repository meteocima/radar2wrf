

#for f in `cd orig; find 08 -name '*.nc' | sort`; do 
#	echo cdo -remapnn,~/xp/2020110900-CAPPI2.nc  orig/$f netcdf/$f.a;
#	echo ncap2 -s 'time=int(time)' netcdf/$f.a netcdf/$f.nc 
#done

for f in `cd orig; find 0{6,7} -name '*.nc' | sort`; do
	if [[ $f == *"CAPPI2"* ]]; then
		VARNAME=CAPPI2
	elif [[ $f == *"CAPPI3"* ]]; then
		VARNAME=CAPPI3
	elif [[ $f == *"CAPPI5"* ]]; then
		VARNAME=CAPPI5
	fi

    INSTANT=${f:3:10}
    Y=${INSTANT:0:4}
    M=${INSTANT:4:2}
    D=${INSTANT:6:2}
    H=${INSTANT:8:2}

	echo cdo -remapnn,~/xp/2020110900-CAPPI2.nc orig/$f netcdf/$f.a;
	echo ncrename -v Band1,$VARNAME netcdf/$f.a  netcdf/$f.b
	echo cdo -settcal,standard -setreftime,1970-1-1,00:00:00,seconds -settaxis,$Y-$M-$D,$H:00:00 netcdf/$f.b netcdf/$f.c
	echo ncap2 -s 'time=int(time)' netcdf/$f.c netcdf/$f
done