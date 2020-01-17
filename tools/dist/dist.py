#!/usr/bin/env python3

import os
import re
import sys
import shutil
import zipfile
import tarfile
import subprocess
from os.path import join as path_join

cfg = {'targets': [('linux', 'amd64', 'tar'),
                   ('windows', '386', 'zip'),
                   ('windows', 'amd64', 'zip'),
                   ('darwin', 'amd64', 'zip')],
       
       'program': 'qtcdbg',
       'src_path': os.path.abspath('../../cmd/qtcdbg'),
       'bin_path_root': os.path.abspath('../../cmd/qtcdbg/dist'),
       'archive_path_root': os.path.abspath('../../arch'),
    }

def get_version():
    os.chdir(cfg['src_path'])
    cmd = ['go', 'build', '.']
    os.system(' '.join(cmd))
    
    cmd = ['qtcdbg', '-v']
    result = subprocess.run(cmd, stdout=subprocess.PIPE)
    version_str = result.stdout.decode('utf-8').rstrip()

    re_ver = re.compile(r'(\d+)\.(\d+)')
    m = re_ver.search(version_str)
    version = (m.group(1), m.group(2))
    
    return version

def copy_archive(archive_filename):
    os.makedirs(cfg['archive_path_root'], exist_ok=True)
    target_path = path_join(cfg['archive_path_root'], archive_filename)
    if os.path.exists(target_path): 
        os.remove(target_path)
    shutil.move(archive_filename, target_path)
    print("created %s" % target_path)

def build_target(t, ver):
    global cfg
    os.chdir(cfg['src_path'])

    build_identifier = "%s-%s" % (t[0], t[1])
    
    out_dir = path_join(cfg['bin_path_root'], build_identifier)

    out_filename = cfg['program']
    if t[0] == 'windows':
        out_filename += '.exe'
    
    out_path = path_join(out_dir, out_filename)

    cmd = ['go', 'build', '-o', out_path, '.' ]
    os.environ['GOOS'] = t[0]
    os.environ['GOARCH'] = t[1]

    print("Building %s/%s to %s" % (t[0], t[1], out_path))
    os.system(' '.join(cmd))

    del os.environ['GOOS']
    del os.environ['GOARCH']

    version_str = '%s.%s' % (ver[0], ver[1])
    
    if t[2] == 'zip':
        os.chdir(out_dir)
        zip_filename = "%s-%s-%s.zip" % (cfg['program'], version_str, build_identifier)
        zipf = zipfile.ZipFile(zip_filename, 'w', zipfile.ZIP_DEFLATED)
        zipf.write(out_filename)
        zipf.close()
        copy_archive(zip_filename)
    elif t[2] == 'tar':
        os.chdir(out_dir)
        tar_filename = "%s-%s-%s.tar.gz" % (cfg['program'], version_str, build_identifier)
        tarf = tarfile.open(tar_filename, "w:gz")
        tarf.add(out_filename)
        tarf.close()
        copy_archive(tar_filename)



if __name__ == '__main__':
    version = get_version()
    
    for target in cfg['targets']:
        build_target(target, version) 
        
