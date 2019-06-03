$1 == "BSS" {
    MAC = $2
    wifi[MAC]["enc"] = "-"
}
$1 == "SSID:" {
    wifi[MAC]["SSID"] = $2
}
$1 == "freq:" {
    wifi[MAC]["freq"] = $NF
}
$1 == "signal:" {
    #wifi[MAC]["sig"] = int ( (100 - $2) / 10) #" " $3
    wifi[MAC]["sig"] = int( (-0.0154*$2*$2)-(0.3794*$2)+98.182)
    if( $2 > -21) { wifi[MAC]["sig"] = 100 }
    if( $2 < -92) { wifi[MAC]["sig"] = 1}
    
}
$1 == "RSN:" {
    wifi[MAC]["enc"] = "WPA"
}
END {
    #printf "%s\t%s\t%s\n","SSID","Signal","Encryption"

    for (w in wifi) {
        printf "%s\t%s\t%s\n",wifi[w]["SSID"],wifi[w]["sig"],wifi[w]["enc"]
    }
}
