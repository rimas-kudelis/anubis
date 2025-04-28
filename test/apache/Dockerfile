FROM httpd:2.4

RUN sed -i \
  -e 's/^#\(LoadModule .*mod_ssl.so\)/\1/' \
  -e 's/^#\(LoadModule .*mod_rewrite.so\)/\1/' \
  -e 's/^#\(LoadModule .*mod_proxy.so\)/\1/' \
  -e 's/^#\(LoadModule .*mod_proxy_http.so\)/\1/' \
  -e 's/^#\(LoadModule .*mod_socache_shmcb.so\)/\1/' \
  -e 's/^#\(LoadModule .*mod_http2.so\)/\1/' \
  conf/httpd.conf
RUN echo '' >> conf/httpd.conf \
  && echo 'IncludeOptional conf.d/*.conf' >> conf/httpd.conf

COPY conf.d ./conf.d
COPY snippets /etc/httpd/snippets