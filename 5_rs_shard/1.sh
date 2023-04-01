#!/bin/bash

count=0  # 总行数

# 递归遍历文件夹下的所有文件和子文件夹
function traverse() {
    for file in $(ls $1)  # 遍历当前文件夹下的所有文件和子文件夹
    do
        if [ -d $1/$file ]  # 如果是子文件夹，递归遍历
        then
            traverse $1/$file
        else
            if [ ${file##*.} = "go" ]  # 如果是后缀名为.go的文件，计算行数
            then
                lines=$(wc -l $1/$file | awk '{print $1}')  # 使用wc命令统计行数
                count=$(($count + $lines))  # 将行数累加到总行数中
            fi
        fi
    done
}

# 调用函数遍历文件夹
traverse $1

echo "Total lines in .go files: $count"
