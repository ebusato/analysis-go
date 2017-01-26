void summary(TString fileName0, TString fileName1="", TString fileName2="", TString fileName3="", TString fileName4="") {
	gStyle->SetOptStat(0);
	gStyle->SetOptTitle(0);
	gStyle->SetPadLeftMargin(0.15);
	gStyle->SetPadRightMargin(0.15);
	gStyle->SetTitleYOffset(1.7);
	
	TChain ch("tree");
	ch.Add(fileName0);
	if(fileName1 != "") {
		ch.Add(fileName1);
	}
	if(fileName2 != "") {
		ch.Add(fileName2);
	}
	if(fileName3 != "") {
		ch.Add(fileName3);
	}
	if(fileName4 != "") {
		ch.Add(fileName4);
	}
	TTreeReader reader(&ch);
        TTreeReaderValue<UInt_t> Run(reader, "Run");
        TTreeReaderValue<UInt_t> Evt(reader, "Evt");
        TTreeReaderArray<UShort_t> IChanAbs240(reader, "IChanAbs240");
        TTreeReaderArray<Double_t> Ampl(reader, "Ampl");
        TTreeReaderArray<Double_t> Charge(reader, "Charge");
	
	TCut c_noSat = "Sat[0] == 0 && Sat[1] == 0";
	TCut c_energyWindow511 = "E[0] > 420 && E[0] < 600 && E[1] > 420 && E[1] < 600";
	TCut c_energyWindow202 = "E[0] > 150 && E[0] < 250 && E[1] > 150 && E[1] < 250";
	TCut c_energyWindow307 = "E[0] > 250 && E[0] < 350 && E[1] > 250 && E[1] < 350";
	
	TCanvas* c1 = new TCanvas("c1", "c1", 1500, 1000);
	c1->Divide(4,3);
	c1->cd(1);
	ch.Draw("IChanAbs240");
	c1->cd(2);
	ch.Draw("IQuartetAbs60");
	c1->cd(3);
	ch.Draw("ILineAbs12");
	c1->cd(4);
	ch.Draw("IQuartetAbs60[1] : IQuartetAbs60[0]>>hQuartet1vsQuartet0(30,0,30,30,30,60)", c_noSat, "colz");
	TH1F* hQuartet1vsQuartet0 = (TH1F*) gDirectory->Get("hQuartet1vsQuartet0");
	hQuartet1vsQuartet0->GetXaxis()->SetTitle("IQuartetAbs60[0]");
	hQuartet1vsQuartet0->GetYaxis()->SetTitle("IQuartetAbs60[1]");
	c1->cd(5);
	ch.Draw("ILineAbs12[1] : ILineAbs12[0]>>hLine1vsLine0(6,0,6,6,6,12)", c_noSat, "colz");
	TH1F* hLine1vsLine0 = (TH1F*) gDirectory->Get("hLine1vsLine0");
	hLine1vsLine0->GetXaxis()->SetTitle("ILineAbs12[0]");
	hLine1vsLine0->GetYaxis()->SetTitle("ILineAbs12[1]");
	c1->cd(6);
	ch.Draw("E[1] : E[0]>>hE1vsE0(200,0,1022,200,0,1022)", "", "colz");
	TH1F* hE1vsE0 = (TH1F*) gDirectory->Get("hE1vsE0");
	hE1vsE0->GetXaxis()->SetTitle("E[0]");
	hE1vsE0->GetYaxis()->SetTitle("E[1]");
	c1->cd(7);
	ch.Draw("E[0]>>hE0(200,0,1100)");
	ch.Draw("E[0]>>hE0noSat(200,0,1100)", "Sat[0] == 0", "same");
	TH1F* hE0noSat = (TH1F*) gDirectory->Get("hE0noSat");
	hE0noSat->SetLineStyle(7);
	ch.Draw("E[1]>>hE1(200,0,1100)", "", "same");
	ch.Draw("E[1]>>hE1noSat(200,0,1100)", "Sat[1] == 0", "same");
	TH1F* hE1 = (TH1F*) gDirectory->Get("hE1");
	TH1F* hE1noSat = (TH1F*) gDirectory->Get("hE1noSat");
	hE1noSat->SetLineStyle(7);
	hE1->SetLineColor(kRed);
	hE1noSat->SetLineColor(kRed);
	c1->cd(8);
	gPad->SetLogy();
	TH1F* hE0 = (TH1F*) gDirectory->Get("hE0");
	hE0->Draw();
	hE0noSat->Draw("same");
	hE1->Draw("same");
	hE1noSat->Draw("same");
	c1->cd(9);
	ch.Draw("T30[0] - T30[1]>>hCRT(200,-30,30)");
	TH1F* hCRT = (TH1F*) gDirectory->Get("hCRT");
	hCRT->Fit("gaus");
	hCRT->SetStats(1);
	c1->cd(10);  
	ch.Draw("T30[0] - T30[1]>>hCRTcuts(200,-10,10)", c_noSat && c_energyWindow511);
	TH1F* hCRTcuts = (TH1F*) gDirectory->Get("hCRTcuts");
	hCRTcuts->Fit("gaus");
	hCRTcuts->SetStats(1);
}