# lukaplaysrv
噜咖播放器airplay开源命令行实现
# 使用说明
需将对应平台下的`ffmpeg`,`ffprobe`放置到执行目录下。
# config.cfg
```
{
    "Name":"Nzlov VR FileServer", //服务器名称
    "FilePath":[    //视频目录集，可指定多个
        {
            "Path":"/Volumes/sys/", //视频目录
            "Sub":false //是否扫描子目录
        }
    ]
}
```
# Protocol

default port : 20066


## /config

e.g:

request:

    http://192.168.1.10:20066/config

response:

    {
        version: 2,
        serverName: "biezhihuadeMacBook-Pro.local",
        serverUuid: "46ff50e1-d3c6-45cb-8015-34b0895673ba",
        address: "http://192.168.1.10:20066",
        listEndpoint: "http://192.168.1.10:20066/list"
    }


## /list

e.g:

request:

   http://192.168.1.10:20066/list

response:

    [
        {
            files: [
                "/s/:Users:biezhihua:Downloads/360%20%C2%B0%20vr%20campus%20girls%20love%20korea%20dance%20%23360vr.mp4",
                "/s/:Users:biezhihua:Downloads/360%20%C2%B0%20vr%20campus%20girls%20love%20korea%20dance%20360vr.mp4",
                "/s/:Users:biezhihua:Downloads/RMVB%E6%B5%8B%E8%AF%95%E6%A0%B7%E5%93%81.rmvb",
                "/s/:Users:biezhihua:Downloads/test.mp4"
            ],
            subtitles: [ "Zootopia.2016.720p.BluRay.x264-SPARKS.cht.srt" ]
        },
        {
            files: [ ],
            subtitles: [ ]
        },
        ...
        ...
    ]


## /metadata/video_path

e.g

request:

    http://192.168.1.10:20066/metadata/s/:Users:biezhihua:Downloads/360%20%C2%B0%20vr%20campus%20girls%20love%20korea%20dance%20%23360vr.mp4

response:

    {
        programs: [ ],
        streams: [
            {
                index: 0,
                codec_name: "h264",
                codec_type: "video",
                width: 3840,
                height: 1920
            },
            {
                index: 1,
                codec_name: "aac",
                codec_type: "audio"
            }
        ],
        format: {
            duration: "140.969000",
            size: "166693408"
        }
    }


## /thumbnail

e.g

request:

    http://192.168.1.10:20066/thumbnail/s/:Users:biezhihua:Downloads/test.mp4.jpg

    note jpg suffix

response:

    image


## /

note `/` used to access the video and subtitle

e.g:

request:

    http://192.168.1.10:20066/s/:Users:biezhihua:Downloads/test.mp4


response:

    play video or get subtitle


