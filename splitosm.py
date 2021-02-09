import yaml
import datetime as dt
import linecache
import subprocess
import threading
import math
import threading

def splitosm( threadno ,startline ,maxline):

    # print message
    print('Start threadno:' ,str(threadno) ,'startline:' ,str(startline) ,'maxline:' ,str(startline+maxline),'\n')
    linecnt ,node = 0 ,0
    
    while True:
        # read a record(node ,way ,relation are valid)
        line = linecache.getline(osmxml, startline+linecnt)
        if line.lstrip(' ').startswith('<node'):
            if line.endswith('/>\n'):
                # ignore no tags record
                linecnt+=1
                continue
            
            node+=1
            while True:
                linecnt+=1
                if linecache.getline(osmxml, startline+linecnt).lstrip(' ').startswith('</node'):
                    #print(str(startline+linecnt))
                    break
        elif line.lstrip(' ').startswith('<way') or line.lstrip(' ').startswith('<relation'):
            # end of process
            break
        linecnt+=1
        # end of element
        if startline+linecnt > startline+maxline:
            break

    # print message
    print('End threadno:' ,str(threadno) ,'endline:' ,str(startline+linecnt))
    print('Elements node:' ,str(node))

if __name__ == '__main__':

    # print message
    print('Start:['+str(dt.datetime.now())+']')

    # open configuration file.
    with open('titan.yaml' ,'r') as yml:
        config = yaml.load(yml, Loader=yaml.SafeLoader)
    # get basic information from yaml
    osmxml = config['osm.xml']
    core = config['machine.core']
    nodeshare = config['core.nodeshare']
    splitnode = config['split.node']
    splitway = config['split.way']
    splitrelation = config['split.relation']

    # get split and thread param
    maxthread = core-1
    linethread = math.ceil(maxline/maxthread)

    # line count in osm.xml
    maxline = int(subprocess.check_output(['wc', '-l', osmxml]).decode().lstrip(' ').split(' ')[0])
    print('maxline:' ,str(maxline))
    # split osm.xml
    subprocess.check_output(['split' ,'-d' ,'-l' ,str(linethread) ,osmxml])

    # exec splitosm thread
    threadcnt=0
    linecnt=1
    threads=[]
    while threadcnt < maxthread:
        th = threading.Thread(target=splitosm ,args=([threadcnt ,linecnt ,linethread]))
        th.start()
        threads.append(th)
        threadcnt+=1
        linecnt+=linethread

    for th in threads:
        th.join()
    # print message
    print('End:['+str(dt.datetime.now())+']')
