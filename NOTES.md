# streamer
exploration of stuff to do with streaming logitech webcam h264 over ws to browser


```
$ v4l2-ctl --list-devices

HD Pro Webcam C920 (usb-0000:00:14.0-5):
        /dev/video0
	
$ ffmpeg -f v4l2 -list_formats all -i /dev/video0

<snip>

[video4linux2,v4l2 @ 0x136dea0] Raw       :     yuyv422 :     YUV 4:2:2 (YUYV) : 640x480 160x90 160x120 176x144 320x180 320x240 352x288 432x240 640x360 800x448 800x600 864x480 960x720 1024x576 1280x720 1600x896 1920x1080 2304x1296 2304x1536
[video4linux2,v4l2 @ 0x136dea0] Compressed:        h264 :                H.264 : 640x480 160x90 160x120 176x144 320x180 320x240 352x288 432x240 640x360 800x448 800x600 864x480 960x720 1024x576 1280x720 1600x896 1920x1080
[video4linux2,v4l2 @ 0x136dea0] Compressed:       mjpeg :                MJPEG : 640x480 160x90 160x120 176x144 320x180 320x240 352x288 432x240 640x360 800x448 800x600 864x480 960x720 1024x576 1280x720 1600x896 1920x1080

```

choose say h264 at 1280 x 720



https://stackoverflow.com/questions/30765700/ffserver-streaming-h-264-from-logitech-c920-without-re-encoding

> Well, it isn't really an answer, but I managed to do this by switching to vlc. Unfortunately I haven't managed to make ffserver to accept the incoming .H264 stream as-is, without re-encoding it and even if I would have, I still would have had this problem of ffmpeg-C920-linux kernel regression: http://sourceforge.net/p/linux-uvc/mailman/message/33164469/
> 
> As such it looked reasonable to abandon the ffmpeg-ffserver line and try vlc.
> 
> In case anybody else is interested, with vlc I managed to achieve the non-reencoded distribution of the C920 webcam's native .H264 feed by running the follow:
> 
> On the Odroid device this will pick up the .H264 stream from the cam and
> 
> streams it via http in mpeg-ts:
> cvlc v4l2:///dev/video0:chroma=h264:width=1920:height=1080 --sout '#standard{access=http,mux=ts,dst=[ip of odroid]:8080,name=stream,mime=video/ts}' -vvv
> 
> On the CentOS 7 server the following takes the stream from the Odroid and multicasts it, so consumers can connect to it, instead of trying to connect to the Odroid device which has a lot more limited bandwidth (wifi):
> 
> vlc http://[ip of odroid]:8080 --sout '#standard{access=http,mux=ts,dst=[ip of centos server]:8080,name=stream,mime=video/ts}' -vvv
> 
> Now I can play this stream in realtime from the VLC player on an device:
> 
> http://[ip of centos server]:8080
> 
> But yes, this isn't really a solution to the original ffmpeg-ffserver problem, but rather a workaround using vlc for the same.
>


MJPEG streaming
https://github.com/defvol/Paparazzo.js

flussonic have an MIT-licensed open source player that uses MSE witih their stream being
video/mp4 with codec avc1.4d401f
audio/mp4 with codex mp4a.40.2


github.com/korandiz/v4l
viewcam demo has about 0.1 seconds latency or less - seems nearly realtime
streamcam demo to ffplay has about 1.5 seconds latency
Ran for over 8 hours without stalling


## MP4 libraries


### ThankYouMotion/mp4/
license MIT
go encoder/decoder that doesn't touch what is in the boxes, but it has the structures for many (all?) boxes in the spec

Can encode

#### Forks with added value:

* andrewzeneski/mp4 is 15 commits ahead and claims to have memory, error and performance fixes (15 commits)
* mstoykov/mp4 is ahead 13 commits again from andrewzeneski/mp4, with further fixes (and some io.Reader vs io.ReadSeeker experiments)

* seifer/mp4 is ahead 28 commits from ThankYouMotion - consider? Timur Nurutdinov made the changes, which is a name seeen in the other two improved versions. ... same changes?

Mstoykov has ftyp box but seifer does not

#### Results of running tests on samples:
Mstoykov can't copy correctly, but gives info (and info present in copy)
Seifer - filter does not compile, cli segfaults on info (invalid memory address or nil pointer reference)
andrewzeneski - filter does not segfaults on info, as for seifer, box types are unknown
ThankYouMotion - go ./... fine with filter, but segfaults on info

#### Identical or cosmetic Forks (or forks of forks)
* jfbus/mp4
* geclaire/mp4
* icewaver/mp4
* hyhy01/mp4 (forked from jfbus/mp4)
* keminar/mp4 (forked from jfbus/mp4)
* piscator/mp4
* sky0014/mp4
* TihsYloH/mp4
* youkebing/mp4

#### Copy-forks that are identical or cosmetic
* mshafiee/mp4
GPL - seems similar to thankyoumotion/mp4
Indeed it is - meld says that these are the same, except that boxes are moved into a subdirectory
The two forks of this project have no changes

#### Forks that are behind
* godeep/mp4


## CMAF

https://www.theoplayer.com/blog/low-latency-chunked-cmaf
> CMAF leverages an ISOBMFF, fMP4 container (ISO/IEC 14496-12:201). 

[moof mdat] repeat

First frame in CMAF segment must be keyframe (IDR)
segment can be chunked, chunk does not need to start with keyframe

Must align keyframes across bitrates; this enables safe bitrate switching at any keyframe

https://www.akamai.com/uk/en/multimedia/documents/white-paper/low-latency-streaming-cmaf-whitepaper.pdf
> In the broadcast world, satellite latencies are in the 3.5-12 s range and terrestrial cable between 
6-12 s,

>With non-chunked encoding, since the Movie Fragment Box (“moof”) must reference all the video samples held in the 
>Media Data Box (“mdat”), an encoder producing non-fragmented mp4 (CMAF) files must wait to encode the last byte 
>of content before it uploads the first byte to the CDN for distribution. This introduces a delay (loss in latency) of one 
>segment duration. In addition, CDNs receiving the incoming segments typically wait to receive the last byte before 
>forwarding on the first byte, and players wait to receive the last byte from the CDN before beginning to decode the 
>first byte. This pattern of repeated accumulation results in an overall latency loss that is an integer multiple of the 
>segment duration. Delays of 5x segment duration are quite common in status quo deployments — ~10 s with 2 s 
>segments and 20 s with 4 s segments— which still lag behind broadcast levels of latency.

Chunk is smallest unit, contains a moof and a mdat atom

one or more chunks forms a fragment

one or more fragments form a segment

a separate header is needed to initialise chunk playback - must pair the behaviour at the client end as well else just get normal latency (presumably this means using MSE to load individual chunks into the sourceBuffer)

One frame per chunk overhead is 100% for audio, 2% for video

Chunked encoding available since 2003 when AVC first standardised

check out dash.js player?  has a low latency option
check out ull-camf with ffmpeg


binary data stuff
https://blog.mgechev.com/2015/02/06/parsing-binary-protocol-data-javascript-typedarrays-blobs/

> var socket = new WebSocket('ws://127.0.0.1:8081');
> socket.binaryType = 'arraybuffer';
> 
> The two possible values for binaryType, of the WebSocket, are arraybuffer and blob. In most cases arraybuffer will be the one, which allows faster processing since it can be used with the synchronous API of the DataView. In case of large pieces of binary data preferable is the blob binary type.
> 
> So how would we process the following example, using DataView and arraybuffer:
> 
> 2      U16      framebuffer-width
> 2      U16      framebuffer-height
> 16 PIXEL_FORMAT server-pixel-format
> 4      U32      name-length
> 
> connection.onmessage = function (e) {
>   var data = e.data;
>   var dv = new DataView(data);
>   var width = dv.getUint16(0);
>   var height = dv.getUint16(2);
>   var format = getPixelFormat(dv);
>   var len = dv.getUint32(20);
>   console.log('We have width: ' + width +
>       'px, height: ' +
>       height + 'px, name length: ' + len);
> };
>



info on fragmented MP4: http://www.ramugedia.com/mp4-container



This reverse proxy uses Mstroykov/mp4 - which doesn't seem to work for me ..  so how does it work for them?
https://github.com/ironsmile/nedomi/tree/master/handler/mp4


Got the contents writing, but the video plays black ... (is the table wrong, given that new version is shorter?)

[tim@w6625 timdrysdale]$ colordiff -y <(xxd sample2.mp4) <(xxd copy2.mp4) | grep mdat 
0001520: 0000 017a 0000 0178 0000 016b 0000 015c  ...z...x... | 0001520: 6461 7400 47f7 456d 6461 7400 0000 0000  dat.G.Emdat
0001640: 6d64 6174 0000 0000 0000 0000 0000 0323  mdat....... | 0001640: 656c 6c69 733d 3220 3878 3864 6374 3d31  ellis=2 8x8

since colordiff is cutting off the ascii (there are 16 ascii per line) we can check individually

[tim@w6625 timdrysdale]$ xxd sample2.mp4 | grep mdat
0001640: 6d64 6174 0000 0000 0000 0000 0000 0323  mdat...........#
[tim@w6625 timdrysdale]$ xxd copy2.mp4 | grep mdat
0001520: 6461 7400 47f7 456d 6461 7400 0000 0000  dat.G.Emdat.....

This suggests we need a pad of 73 chars, plus-minus fencepost.

Could try hardcoding ...

Ok so it turns out we need a padding of .... 281 bytes for sample2.mp4
and then it plays in VLC.


## H264 stuff
Information on NALU in h264
https://stackoverflow.com/questions/24884827/possible-locations-for-sequence-picture-parameter-sets-for-h-264-stream/24890903#24890903
	



v4l2-ctl --list-formats-ext[tim@w6625 timdrysdale]$ v4l2-ctl --list-formats-ext
ioctl: VIDIOC_ENUM_FMT
        Index       : 0
        Type        : Video Capture
        Pixel Format: 'YUYV'
        Name        : YUV 4:2:2 (YUYV)
                Size: Discrete 640x480
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 160x90
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 160x120
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 176x144
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 320x180
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 320x240
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 352x288
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 432x240
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 640x360
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 800x448
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 800x600
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 864x480
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 960x720
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 1024x576
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 1280x720
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 1600x896
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 1920x1080
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 2304x1296
                        Interval: Discrete 0.500s (2.000 fps)
                Size: Discrete 2304x1536
                        Interval: Discrete 0.500s (2.000 fps)

        Index       : 1
        Type        : Video Capture
        Pixel Format: 'H264' (compressed)
        Name        : H.264
                Size: Discrete 640x480
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 160x90
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 160x120
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 176x144
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 320x180
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 320x240
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 352x288
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 432x240
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 640x360
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 800x448
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 800x600
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 864x480
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 960x720
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 1024x576
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 1280x720
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 1600x896
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 1920x1080
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)

        Index       : 2
        Type        : Video Capture
        Pixel Format: 'MJPG' (compressed)
        Name        : MJPEG
                Size: Discrete 640x480
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 160x90
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 160x120
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 176x144
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 320x180
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 320x240
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 352x288
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 432x240
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 640x360
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 800x448
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 800x600
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 864x480
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 960x720
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 1024x576
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 1280x720
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 1600x896
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)
                Size: Discrete 1920x1080
                        Interval: Discrete 0.033s (30.000 fps)
                        Interval: Discrete 0.042s (24.000 fps)
                        Interval: Discrete 0.050s (20.000 fps)
                        Interval: Discrete 0.067s (15.000 fps)
                        Interval: Discrete 0.100s (10.000 fps)
                        Interval: Discrete 0.133s (7.500 fps)
                        Interval: Discrete 0.200s (5.000 fps)

So if use YUV then limited frame rate for large resolutions.

viewcamera demo uses YUV

no support for h264 in viewcam demo

some info here on ffmpeg with h264?
http://dranger.com/ffmpeg/tutorial01.html

https://godoc.org/github.com/prohulaelk/codec
https://godoc.org/github.com/giorgisio/goav/avcodec

FourCC for H264
'X','2','6','4'
of H 2 6 4 ...

streamcam written to file is playing in vlc

https://stackoverflow.com/questions/6756770/libavcodec-how-to-tell-end-of-access-unit-when-decoding-h-264-stream
> I don't know what is your RTP source, I have been dealing with IP cameras which - even those professional ones - encode H.264 in a quite straightforwrd way: I-Frame: [optional NALU 7], [optional NALU 8], NALU 5, [optional NALU 6] P-Frame: NALU 1, [optional NALU 6] This lets you break them into frames easily. – Roman R. Sep 7 '11 at 18:34


Putting it into mp4 format

https://stackoverflow.com/questions/53992918/about-fmp4-encoding-how-to-fill-the-mdat-box-wit-h264-frame


> 
> H.264 can be in different stream formats. One is called "Annex B" the other one is MP4. In "Annex B" your NAL units are prefix with start codes 00 00 00 01 or 00 00 01. In MP4 your NAL units are prefixed with the size in bytes.
> 
> I assume your encoder emits "Annex B".
> 
>     Remove the start code (00) 00 00 01
> 
>     Prefix your NAL units with the size (typically 4 bytes)
> 
>     Filter out AUD/SPS/PPS NAL units from your stream
> 
>     Write you converted NAL units into the MDAT box
> 
>     Create an AVC Configuration Box ('avcC') based on your SPS, PPS and the length size
> 
>     Store your avcC box in moov->trak->mdia->minf->stbl->avc1->avcC
> 
>     While you are writing your samples into mdat - keep track of sizes, offsets and frame types to create the correct stts, stss, stsc, stsz and stco boxes.
> 
>

https://github.com/axiomatic-systems/Bento4

install with make, put executables in ~/bin, put bin on path (no 'install' target in makefile, and scons needs py2)



https://github.com/hamishcoleman/h264mux
some perl for streaming h264 of raspberry pi - or workups toward that solution.

for working with ffmpeg to produce mp4 frags
kevinGodell/mp4frag

git clone <repo>
cd <repo>
npm install
node ./tests/test.js

no libx264 in my installation...
ffmpeg -f v4l2 -framerate 25 -video_size 800x600 -i /dev/video0 -c:v libx264 -f mp4 -movflags +dash pipe:1

but docker image does
sudo docker run --device=/dev/video0 jrottenberg/ffmpeg -f v4l2 -framerate 25 -video_size 800x600 -i /dev/video0 -c:v libx264 -f mp4 -movflags +dash pipe:1


v4l2-ctl --device /dev/video0 --set-fmt-video=width=800,height=600,pixelformat=H264





### Streaming like 1999 ...


https://github.com/phoboslab/jsmpeg

3 days of streaming
20% CPU, 0.1% of memory on Thnkstation

stream a video to the server instead of the camera
ffmpeg -f mpegts -i ~/Downloads/20120927_085129.m2ts -f mpegts -codec:v mpeg1video -s 640x480 -b:v 1000k -bf 0 http://localhost:8081/supersecret

can run ffmpeg on AWS and stream an m2ts file to thinkstation on uni net no worries, when streaming to the localhost port on AWS

So ... now need a relay that logs into the local websocket relay, and transmits to the remote websocket relay


audio ...
ffmpeg -f alsa -ar 44100 -c 2 -i hw:0 -f mpegts -codec:a mp2 -b:a 128k -muxdelay 0.001 http://localhost:8081/supersecret

./encode_video.sh

## identifying the webcam

logitech devices are reported to have distinct IP addresses obtainable with

```sudo lsusb -v -d 046d:082d | grep -i serial
  	iSerial                 1 702F5F4F

```

(See https://superuser.com/questions/902012/how-to-identify-usb-webcam-by-serial-number-from-the-linux-command-line)


v4l2-ctl --list-devices

Even better is:
```
$ ls /dev/v4l/by-id
usb-046d_HD_Pro_Webcam_C920_42B15FEF-video-index0  usb-046d_HD_Pro_Webcam_C920_702F5F4F-video-index0
```

so this then works:
```
$ ffmpeg -f v4l2 -framerate 25 -video_size 640x480 -i /dev/v4l/by-id/usb-046d_HD_Pro_Webcam_C920_42B15FEF-video-index0  -f mpegts -codec:v mpeg1video -s 640x480 -b:v 1000k -bf 0 http://localhost:8081/supersecret
```

## Drone stack webrtc
https://webrtchacks.com/what-i-learned-about-h-264-for-webrtc-video-tim-panton/

## MPEG TS buffering and delays
https://tools.ietf.org/html/draft-begen-avt-rtp-mpeg2ts-preamble-06

## fMP4
https://tools.ietf.org/html/rfc8216#page-7
https://www.iso.org/standard/68960.html

mux issues
https://github.com/videojs/mux.js/issues/144


### chat room

this is accessible from behind the firewall
```
$ wscat --connect ws://jsmpeg.practable.io:80/ws/time/
connected (press CTRL+C to quit)
< 2019-08-12 00:22:49.599484002 +0000 UTC m=+50.002033542
< 2019-08-12 00:22:59.599415086 +0000 UTC m=+60.001964625
< 2019-08-12 00:23:09.599424551 +0000 UTC m=+70.001974087
< 2019-08-12 00:23:19.599433096 +0000 UTC m=+80.001982650
< 2019-08-12 00:23:29.599428148 +0000 UTC m=+90.001977696
```

can connect to the AWS relay and send and receive messages
wscat --connect  ws://jsmpeg.practable.io:80/ws/camera/

can connect to the local relay and see camera data
wscat --connect  ws://localhost:8082/


## Browser issues

https://github.com/kevinGodell/mp4frag/issues/1
Need to change source buffer mode from sequence to segment to be able to start anywhere in a stream in chrome


## Twitch
https://blog.twitch.tv/live-video-transmuxing-transcoding-ffmpeg-vs-twitchtranscoder-part-i-489c1c125f28


## js JS players

mp4: https://github.com/mbebenita/Broadway
 (needs to ingest whole film first)
 

## ChihChengYang and jsmpegs https2ws

stream h264
this did not work (segfault)
ffmpeg -y -s 800x600 -f video4linux2 -i /dev/video0 -c:v libx264 -tune zerolatency -an http://localhost:8081/supersecret

note this criticism of c920's h264 here:https://unix.stackexchange.com/questions/163977/ffserver-streaming-h-264-from-logitech-c920

v4l2-ctl --set-fmt-video=width=800,height=600,pixelformat=H264

still segfaults even after setting format (just manages a few frames)

this seems to stream:-
sudo docker run --network="host" --device=/dev/video0 jrottenberg/ffmpeg -f v4l2 -framerate 25 -video_size 640x480 -i /dev/video0 -codec:v libx264 -f h264 http://localhost:8081/supersecret


#Useful stuff

Mime type support checking on browser
https://cconcolato.github.io/media-mime-support/


grep -rnw './' -e 'Decode'

colordiff -y <(xxd sample2.mp4) <(xxd copy2.mp4)
colordiff -y <(xxd -l1000 sample2.mp4) <(xxd -l1000 copy2.mp4)

getting node to run as sudo
n=$(which node); \
n=${n%/bin/node}; \
chmod -R 755 $n/bin/*; \
sudo cp -r $n/{bin,lib,share} /usr/local

except can't run that due to 'dangling symlinks'


