import yaml
import datetime as dt
import linecache
import subprocess
import threading
import math

def splitosm( threadno ,startline ,line):

    # print message
    print('Start threadno:' ,str(threadno) ,'startline:' ,str(startline) ,'maxline:' ,str(startline+line))
    linecnt ,node ,way ,relation ,inloop = 0 ,0 ,0 ,0 ,True
    elcontinue = ""
    while inloop:
        # read a record(node ,way ,relation are valid)
        line = linecache.getline(osmxml, startline))
        if line.lstrip(' ').startswith('<node') or elcontinue='node':
            # convert id ,lat ,lon　to CSV object
            # 
            if line.lstrip(' ').endswith('\/>') or line.lstrip(' ').startswith('<node'):
                # end of element
                node+=1
                elcontinue=''
            else:
                # continue element
                elcontinue='node'
        elif line.lstrip(' ').startswith('<way') or elcontinue='way':
            way+=1                
        elif line.lstrip(' ').startswith('<relation') or elcontinue='relation':
            relation+=1
        linecnt+=1

    # print message
    print('End threadno:' ,str(threadno) ,'endline:' ,str(startline+line))
    print('Elements node:' ,str(node) ,'way:' ,str(way),'relation:' ,str(relation))

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

    # line count in osm.xml
    maxline = int(subprocess.check_output(['wc', '-l', osmxml]).decode().lstrip(' ').split(' ')[0])
    print('maxline:' ,str(maxline))
    # cacheing and read osm.xml
    dummy = linecache.getline(osmxml, 0)
    # get thread param
    maxthread = core-1
    linethread = math.ceil(maxline/maxthread)

    # exec splitosm thread
    threadcnt=0
    linecnt=1
    while threadcnt < maxthread:
        splitosm(threadcnt ,linecnt ,linethread)
        threadcnt+=1
        linecnt+=linethread

    # print message
    print('End:['+str(dt.datetime.now())+']')
