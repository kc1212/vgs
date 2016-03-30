execs = cli discosrv gridsdr resman

.PHONY: clean $(execs)

all: $(execs)

$(execs):
	go build -v github.com/kc1212/vgs/cmd/$@

clean:
	rm -f $(execs)
