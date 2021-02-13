import yaml
import datetime as dt
import linecache
import subprocess
import threading
import math
import threading
import glob
import os

def readosmxml( threadno ,filename):

    # print message
    print('Start threadno:' ,str(threadno) ,'filename:' ,filename,'\n')
    node = 0 
    
    with open(filename) as f:
        # read a record(node ,way ,relation are valid)
        for line in f:
            if line.lstrip(' ').startswith('<node'):
                if line.endswith('/>\n'):
                    # ignore no tags record
                    continue
                node+=1
            elif line.lstrip(' ').startswith('<way') or line.lstrip(' ').startswith('<relation'):
                # end of process
                break

    # print message
    print('End threadno:' ,str(threadno) ,'Elements node:' ,str(node))

def alpha2num(alpha):
    num=0
    for index, item in enumerate(list(alpha)):
        num += pow(26,len(alpha)-index-1)*(ord(item)-ord('a')+1)
    return num

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

    # get size for osm.xml and split files
    osmxmlsize = os.path.getsize(osmxml)
    maxthread = core-1
    size = math.ceil(osmxmlsize/maxthread)
    print('osmxmlsize:' ,str(osmxmlsize) ,'size:' ,str(size))
    # remove split files and split osm.xml
    for f in glob.glob(osmxml+'.*'):
        os.remove(f)
    subprocess.check_output(['split' ,'-b' ,str(size) ,osmxml ,osmxml+'.'])
    # get split osmxml(name sort)
    osmxmlfiles = sorted(glob.glob(osmxml+'.*'))
    splitlist = []
    for filename in osmxmlfiles:
        # extent convert with number(extent start with 'aa')
        ext = os.path.splitext(filename)[1].lstrip('.')
        splitfilename = os.path.splitext(os.path.basename(filename))[0]+'.'+str(alpha2num(ext) - 27)
        os.rename(filename ,splitfilename)
        splitlist.append(splitfilename)

    # exec splitosm thread
    threadcnt=0
    threads=[]
    while threadcnt < maxthread:
        th = threading.Thread(target=readosmxml ,args=([threadcnt ,splitlist[threadcnt]]))
        th.start()
        threads.append(th)
        threadcnt+=1

    for th in threads:
        th.join()
    # print message
    print('End:['+str(dt.datetime.now())+']')
