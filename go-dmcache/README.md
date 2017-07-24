dm-cache Notes
==============

Caveats when Using lvmcache
---------------------------
- Volume(s) to be cached must reside in the same VG as the cachepool. For
  example:
```
# lvconvert --type cache /dev/vg01/data01 --cachepool /dev/vg02/lv_cache
   Please use a single volume group name ("vg01" or "vg02")
```

- Each volume to be cached requires its own cachepool (i.e., cache + metadata
  pair). For example:
```
# lvconvert --type cache /dev/vg01/data01 --cachepool /dev/vg01/lv_cache
   Logical volume vg01/data01 is now cached.
# lvconvert --type cache /dev/vg01/data02 --cachepool /dev/vg01/lv_cache
   lv_cache is already in use by data01
```

lvmcache HOWTO
--------------

1. Create volume group that spans both fast and slow storage devices:
```
# vgcreate vg01 /dev/vdb1 /dev/vdc1
```

2. Create two test-data volumes on specific PV (slow) targets:
```
# lvcreate -L 4g -n data01 vg01 /dev/vdb1
# lvcreate -L 4g -n data02 vg01 /dev/vdb1
```

3. Create cache volumes on specific PV (fast) targets:
```
# lvcreate -L 500m -n lv_cache_meta vg01 /dev/vdc1
# lvcreate -L 9G -n lv_cache vg01 /dev/vdc1
```

4. Associate lv_cache_meta with lv_cache
```
# lvconvert --type cache-pool --cachemode writeback /dev/vg01/lv_cache \
  --poolmetadata /dev/vg01/lv_cache_meta --chunksize 256
```

5. Cache the data volume
```
# lvconvert --type cache /dev/vg01/data01 --cachepool /dev/vg01/lv_cache
```

The /dev/vg01/data01 volume should now be cached. Verify with `dmsetup`:
```
# dmsetup status /dev/vg01/data01
0 8388608 cache 8 86/128000 512 784/36864 200 67 1018 0 0 784 0 1 writeback 2 migration_threshold 2048 smq 0 rw -
```

```
# lvs
  LV       VG   Attr       LSize Pool     Origin         Data%  Meta%  Move Log Cpy%Sync Convert
  data01   vg01 Cwi-aoC--- 4.00g lv_cache [data01_corig]                                        
  data02   vg01 -wi-a----- 4.00g                                                                
  lv_cache vg01 Cwi---C--- 9.00g
```
