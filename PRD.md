https://data.people.com.cn/




```
https://data.people.com.cn/rmrb/s?type=2&qs=%7B%22cds%22%3A%5B%7B%22fld%22%3A%22dataTime.start%22%2C%22cdr%22%3A%22AND%22%2C%22hlt%22%3A%22false%22%2C%22vlr%22%3A%22AND%22%2C%22qtp%22%3A%22DEF%22%2C%22val%22%3A%222025-08-07%22%7D%2C%7B%22fld%22%3A%22dataTime.end%22%2C%22cdr%22%3A%22AND%22%2C%22hlt%22%3A%22false%22%2C%22vlr%22%3A%22AND%22%2C%22qtp%22%3A%22DEF%22%2C%22val%22%3A%222025-08-29%22%7D%5D%2C%22obs%22%3A%5B%7B%22fld%22%3A%22dataTime%22%2C%22drt%22%3A%22DESC%22%7D%5D%7D
```

URLdecoded
->

```
https://data.people.com.cn/rmrb/s?type=2&qs={"cds":[{"fld":"dataTime.start","cdr":"AND","hlt":"false","vlr":"AND","qtp":"DEF","val":"2025-08-07"},{"fld":"dataTime.end","cdr":"AND","hlt":"false","vlr":"AND","qtp":"DEF","val":"2025-08-29"}],"obs":[{"fld":"dataTime","drt":"DESC"}]}
```

```
https://data.people.com.cn/rmrb/s

type=2

qs=
{
  "cds": [
    {
      "fld": "dataTime.start",
      "cdr": "AND",
      "hlt": "false",
      "vlr": "AND",
      "qtp": "DEF",
      "val": "2025-08-07"
    },
    {
      "fld": "dataTime.end",
      "cdr": "AND",
      "hlt": "false",
      "vlr": "AND",
      "qtp": "DEF",
      "val": "2025-08-29"
    }
  ],
  "obs": [
    {
      "fld": "dataTime",
      # 倒序
      "drt": "DESC"
    }
  ]
}


```

## 背景

人民日报网站 https://data.people.com.cn/
有很多文章，我们需要去爬取这些文章

## 任务拆解
我们发现可以直接定位文章
基于开始日期，结束日期(查询区间设置一个月)
[1949到2025]

pageNo 从1开始
position 0开始19

```
https://data.people.com.cn/rmrb/s?type=2&qs=%7B%22cds%22%3A%5B%7B%22fld%22%3A%22dataTime.start%22%2C%22cdr%22%3A%22AND%22%2C%22hlt%22%3A%22false%22%2C%22vlr%22%3A%22AND%22%2C%22qtp%22%3A%22DEF%22%2C%22val%22%3A%222025-08-01%22%7D%2C%7B%22fld%22%3A%22dataTime.end%22%2C%22cdr%22%3A%22AND%22%2C%22hlt%22%3A%22false%22%2C%22vlr%22%3A%22AND%22%2C%22qtp%22%3A%22DEF%22%2C%22val%22%3A%222025-08-29%22%7D%5D%2C%22obs%22%3A%5B%7B%22fld%22%3A%22dataTime%22%2C%22drt%22%3A%22DESC%22%7D%5D%7D


https://data.people.com.cn/rmrb/pd.html?qs={"cds":[{"fld":"dataTime.start","cdr":"AND","hlt":"false","vlr":"AND","qtp":"DEF","val":"2025-08-01"},{"fld":"dataTime.end","cdr":"AND","hlt":"false","vlr":"AND","qtp":"DEF","val":"2025-08-29"}],"obs":[{"fld":"dataTime","drt":"DESC"}]}&tr=A&pageNo=1&pageSize=20&position=0
```


