import yaml
import datetime as dt
import linecache
import subprocess
from multiprocessing import Process
import math
import glob
import os
import csv

def readosmxml(osmxml ,procno ,placevalues):

    # print message
    print('Start procno:' ,str(procno) ,'filename:' ,osmxml,'\n')
    nodecnt ,tagflag = 0 ,False
    csvlist = []
    
    with open(osmxml+'.'+str(procno)) as f:
        # read a record(node ,way ,relation are valid)
        for line in f:
            if line.lstrip(' ').startswith('<node') and line.lstrip(' ').endswith('/>\n'):
                # ignore no tags record
                continue                
            if line.lstrip(' ').startswith('<node') or tagflag:
                # about node
                if tagflag==False:
                    # start on <node..
                    tagflag = True
                    # 次のファイル読み込み実装までの一時的な判定
                    if line.lstrip(' ').startswith('<node') and line.lstrip(' ').endswith('\n'):
                        node = line.lstrip(' ').split(' ')
                        id = node[1].lstrip('id=').replace('\"','')
                        lat = str(node[4].lstrip('lat=').replace('\"',''))
                        lon = str(node[5].lstrip('lon=').replace('\"','')).replace('/','').replace('>\n','')
                    # clear tag inf
                    taginf={}
                    nodecnt+=1
                else:
                    # 次のファイル読み込み実装までの一時的な判定
                    if line.lstrip(' ').endswith('\n'):
                        # <tag.. or </node..
                        if line.lstrip(' ').startswith('<tag'):
                            # get tag data
                            tag = line.lstrip(' ').split(' ')
                            taginf[tag[1].lstrip('k=').replace('\"','')] = tag[2].lstrip('v=').replace('\"','').replace('/>\n','')
                        elif line.lstrip(' ').startswith('</node>\n'):
                            # end of <node process
                            if 'place' in taginf:
                                # yamlからのリストによる判定実装までの一時的な処置
                                if taginf['place']=='city' or taginf['place']=='town' or taginf['place']=='village' or taginf['place']=='island' or taginf['place']=='quarter' or taginf['place']=='neighbourhood' or taginf['place']=='locality':
                                    value = [id,taginf.get('place'),taginf.get('name'),taginf.get('name:ja-Hira'),taginf.get('name:en'),lon,lat]
                                    csvlist.append(value)
                            tagflag=False
            elif line.lstrip(' ').startswith('<way') or line.lstrip(' ').startswith('<relation'):
                # end of process
                break
    # output csv
    with open(osmxml+'.'+str(procno)+'.csv' ,'w') as csvf:
        csvwriter = csv.writer(csvf)
        if procno==0:
            # write header
            csvwriter.writerow(['#id ,place ,name ,name:ja-Hira ,name:en ,lon ,lat'] ,)
        # write data
        csvwriter.writerows(csvlist)
    # print message
    print('End:['+str(dt.datetime.now())+']'+' procno:' ,str(procno) ,'Elements node:' ,str(nodecnt))

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
    core = config['core']
    placevalues = config['placevalues']
    
    # get size for osm.xml and split files
    osmxmlsize = os.path.getsize(osmxml)
    maxproc = core-1 if core-1 > 0 else 1 
    size = math.ceil(osmxmlsize/maxproc)
    print('osmxmlsize:' ,str(osmxmlsize) ,'size:' ,str(size))
    # remove split files and split osm.xml
    for f in glob.glob(osmxml+'.*'):
        os.remove(f)
    for f in glob.glob(osmxml+'.*.*'):
        os.remove(f)
    subprocess.check_output(['split' ,'-b' ,str(size) ,osmxml ,osmxml+'.'])
    # get split osmxml(name sort)
    osmxmlfiles = sorted(glob.glob(osmxml+'.*'))
    for filename in osmxmlfiles:
        # extent convert with number(convert '.aa' to '.0')
        ext = os.path.splitext(filename)[1].lstrip('.')
        splitfilename = os.path.splitext(os.path.basename(filename))[0]+'.'+str(alpha2num(ext) - 27)
        os.rename(filename ,splitfilename)

    print('proc Start:['+str(dt.datetime.now())+']')
    # exec splitosm proc
    procno ,proclist = 0 ,[]
    while procno < maxproc:
        proc = Process(target=readosmxml ,args=([osmxml ,procno ,placevalues]))
        proc.start()
        proclist.append(proc)
        procno+=1
    # waiting for end of process
    for proc in proclist:
        proc.join()
    csvlist = sorted(glob.glob(osmxml+'.*.csv'))
    command = ['cat']
    command = command + csvlist
    command.append('>'+osmxml+'.csv')
    commandstr = ' '.join(command)
    subprocess.check_output(commandstr ,shell=True)
    # print message
    print('End:['+str(dt.datetime.now())+']')
