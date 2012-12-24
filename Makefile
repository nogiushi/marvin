MARVIN_LESS = ./less/marvin.less


all:  bootstrap jquery

clean:
	rm -r static/bootstrap
	rm -r static/jquery

bootstrap:
	mkdir -p static
	cp -r components/bootstrap/bootstrap static/

jquery:
	mkdir -p static
	cp -r components/jquery static/


