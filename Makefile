MARVIN_LESS = ./less/marvin.less


all:  bower bootstrap jquery

clean:
	rm -r static/bootstrap
	rm -r static/jquery

bower:
	bower install
	cd components/bootstrap/; make; make bootstrap

bootstrap:
	mkdir -p static
	cp -r components/bootstrap/bootstrap static/

jquery:
	mkdir -p static
	cp -r components/jquery static/


